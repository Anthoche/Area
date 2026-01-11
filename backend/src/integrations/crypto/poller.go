package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"area/src/workflows"
)

const defaultInterval = 2 * time.Minute

// StartCryptoPoller checks crypto workflows and triggers on price or change thresholds.
func StartCryptoPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("crypto poller: missing dependencies, skipping")
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		lastPriceState := make(map[int64]bool)
		lastChangeState := make(map[int64]bool)
		lastCheckPrice := make(map[int64]time.Time)
		lastCheckChange := make(map[int64]time.Time)

		for {
			select {
			case <-ctx.Done():
				log.Printf("crypto poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			pollPriceThreshold(ctx, wfStore, wfService, lastPriceState, lastCheckPrice)
			pollPercentChange(ctx, wfStore, wfService, lastChangeState, lastCheckChange)
		}
	}()
}

// pollPriceThreshold checks price threshold workflows and triggers on crossings.
func pollPriceThreshold(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, lastState map[int64]bool, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "crypto_price_threshold")
	if err != nil {
		log.Printf("crypto poller: list price workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}

		cfg, err := workflows.CryptoPriceThresholdConfigFromJSON(wf.TriggerConfig)
		if err != nil || strings.TrimSpace(cfg.CoinID) == "" {
			log.Printf("crypto poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if last, ok := lastCheck[wf.ID]; ok && time.Since(last) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		coin, err := fetchMarket(ctx, cfg.CoinID, cfg.Currency)
		if err != nil {
			log.Printf("crypto poller wf %d: market: %v", wf.ID, err)
			continue
		}

		above := coin.Price >= cfg.Threshold
		prev, hasPrev := lastState[wf.ID]
		lastState[wf.ID] = above
		if !hasPrev {
			payload := buildPricePayload(cfg, coin, "current")
			if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
				log.Printf("crypto poller trigger wf %d: %v", wf.ID, err)
			}
			continue
		}

		dir := strings.ToLower(strings.TrimSpace(cfg.Direction))
		trigger := (dir == "above" && above && !prev) || (dir == "below" && !above && prev)
		if !trigger {
			continue
		}
		payload := buildPricePayload(cfg, coin, "threshold")
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("crypto poller trigger wf %d: %v", wf.ID, err)
		}
	}
}

// pollPercentChange checks percent change workflows and triggers on threshold crossings.
func pollPercentChange(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, lastState map[int64]bool, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "crypto_percent_change")
	if err != nil {
		log.Printf("crypto poller: list change workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}

		cfg, err := workflows.CryptoPercentChangeConfigFromJSON(wf.TriggerConfig)
		if err != nil || strings.TrimSpace(cfg.CoinID) == "" {
			log.Printf("crypto poller wf %d: bad config: %v", wf.ID, err)
			continue
		}
		interval := defaultInterval
		if cfg.IntervalMin > 0 {
			interval = time.Duration(cfg.IntervalMin) * time.Minute
		}
		if last, ok := lastCheck[wf.ID]; ok && time.Since(last) < interval {
			continue
		}
		lastCheck[wf.ID] = time.Now()

		coin, err := fetchMarket(ctx, cfg.CoinID, cfg.Currency)
		if err != nil {
			log.Printf("crypto poller wf %d: market: %v", wf.ID, err)
			continue
		}
		change := coin.Change1H
		if cfg.Period == "24h" {
			change = coin.Change24H
		}
		state := evaluateChange(change, cfg.Percent, cfg.Direction)
		prev, hasPrev := lastState[wf.ID]
		lastState[wf.ID] = state
		if !hasPrev {
			if state {
				payload := buildChangePayload(cfg, coin, change, "current")
				if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
					log.Printf("crypto poller trigger wf %d: %v", wf.ID, err)
				}
			}
			continue
		}
		if !state || prev {
			continue
		}
		payload := buildChangePayload(cfg, coin, change, "threshold")
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("crypto poller trigger wf %d: %v", wf.ID, err)
		}
	}
}

// evaluateChange checks if the change meets the threshold in the specified direction.
func evaluateChange(value, threshold float64, direction string) bool {
	dir := strings.ToLower(strings.TrimSpace(direction))
	switch dir {
	case "above":
		return value >= threshold
	case "below":
		return value <= -threshold
	default:
		if value < 0 {
			value = -value
		}
		return value >= threshold
	}
}

// buildPricePayload constructs the payload for price threshold events.
func buildPricePayload(cfg workflows.CryptoPriceThresholdConfig, coin *marketCoin, event string) map[string]any {
	payload := map[string]any{
		"coin_id":    coin.ID,
		"symbol":     coin.Symbol,
		"name":       coin.Name,
		"currency":   coin.Currency,
		"price":      coin.Price,
		"threshold":  cfg.Threshold,
		"direction":  cfg.Direction,
		"event":      event,
		"timestamp":  time.Now().Format(time.RFC3339),
		"change_1h":  coin.Change1H,
		"change_24h": coin.Change24H,
	}
	for k, v := range cfg.PayloadTemplate {
		payload[k] = v
	}
	if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
		payload["content"] = fmt.Sprintf("%s | %.4f %s", content, coin.Price, strings.ToUpper(coin.Currency))
	} else {
		payload["content"] = fmt.Sprintf("%s price: %.4f %s", strings.ToUpper(coin.Symbol), coin.Price, strings.ToUpper(coin.Currency))
	}
	return payload
}

// buildChangePayload constructs the payload for percent change events.
func buildChangePayload(cfg workflows.CryptoPercentChangeConfig, coin *marketCoin, change float64, event string) map[string]any {
	payload := map[string]any{
		"coin_id":    coin.ID,
		"symbol":     coin.Symbol,
		"name":       coin.Name,
		"currency":   coin.Currency,
		"price":      coin.Price,
		"period":     cfg.Period,
		"percent":    cfg.Percent,
		"direction":  cfg.Direction,
		"change":     change,
		"event":      event,
		"timestamp":  time.Now().Format(time.RFC3339),
		"change_1h":  coin.Change1H,
		"change_24h": coin.Change24H,
	}
	for k, v := range cfg.PayloadTemplate {
		payload[k] = v
	}
	if content, ok := payload["content"]; ok && fmt.Sprint(content) != "" {
		payload["content"] = fmt.Sprintf("%s | %s change: %.2f%%", content, cfg.Period, change)
	} else {
		payload["content"] = fmt.Sprintf("%s %s change: %.2f%%", strings.ToUpper(coin.Symbol), cfg.Period, change)
	}
	return payload
}

type marketCoin struct {
	ID        string
	Symbol    string
	Name      string
	Price     float64
	Currency  string
	Change1H  float64
	Change24H float64
}

// fetchMarket retrieves market data for a coin from CoinGecko.
func fetchMarket(ctx context.Context, coinID, currency string) (*marketCoin, error) {
	id := strings.ToLower(strings.TrimSpace(coinID))
	if id == "" {
		return nil, fmt.Errorf("missing coin_id")
	}
	cc := strings.ToLower(strings.TrimSpace(currency))
	if cc == "" {
		cc = "usd"
	}
	u := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=%s&ids=%s&price_change_percentage=1h,24h",
		url.QueryEscape(cc),
		url.QueryEscape(id),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "area-app/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coingecko status %d: %s", resp.StatusCode, string(body))
	}
	var raw []struct {
		ID        string  `json:"id"`
		Symbol    string  `json:"symbol"`
		Name      string  `json:"name"`
		Price     float64 `json:"current_price"`
		Change1H  float64 `json:"price_change_percentage_1h_in_currency"`
		Change24H float64 `json:"price_change_percentage_24h_in_currency"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("coin not found")
	}
	return &marketCoin{
		ID:        raw[0].ID,
		Symbol:    raw[0].Symbol,
		Name:      raw[0].Name,
		Price:     raw[0].Price,
		Currency:  cc,
		Change1H:  raw[0].Change1H,
		Change24H: raw[0].Change24H,
	}, nil
}
