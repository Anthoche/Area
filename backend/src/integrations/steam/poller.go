package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"area/src/workflows"
)

const defaultInterval = 2 * time.Minute

// StartSteamPoller checks steam workflows and triggers on status or price changes.
func StartSteamPoller(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service) {
	if wfStore == nil || wfService == nil {
		log.Println("steam poller: missing dependencies, skipping")
		return
	}
	apiKey := strings.TrimSpace(os.Getenv("STEAM_API_KEY"))
	if apiKey == "" {
		log.Println("steam poller: missing STEAM_API_KEY, player status disabled")
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		lastOnline := make(map[int64]int)
		lastSale := make(map[int64]int)
		lastPrice := make(map[int64]int)
		lastCheckOnline := make(map[int64]time.Time)
		lastCheckSale := make(map[int64]time.Time)
		lastCheckPrice := make(map[int64]time.Time)

		for {
			select {
			case <-ctx.Done():
				log.Printf("steam poller stopped: %v", ctx.Err())
				return
			case <-ticker.C:
			}

			pollPlayerOnline(ctx, wfStore, wfService, apiKey, lastOnline, lastCheckOnline)
			pollGameSales(ctx, wfStore, wfService, lastSale, lastCheckSale)
			pollPriceChanges(ctx, wfStore, wfService, lastPrice, lastCheckPrice)
		}
	}()
}

func pollPlayerOnline(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, apiKey string, lastOnline map[int64]int, lastCheck map[int64]time.Time) {
	if strings.TrimSpace(apiKey) == "" {
		return
	}
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "steam_player_online")
	if err != nil {
		log.Printf("steam poller: list player workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.SteamPlayerOnlineConfigFromJSON(wf.TriggerConfig)
		if err != nil || strings.TrimSpace(cfg.SteamID) == "" {
			log.Printf("steam poller wf %d: bad config: %v", wf.ID, err)
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

		player, err := fetchPlayerSummary(ctx, apiKey, cfg.SteamID)
		if err != nil {
			log.Printf("steam poller wf %d: player summary: %v", wf.ID, err)
			continue
		}

		prev, ok := lastOnline[wf.ID]
		lastOnline[wf.ID] = player.PersonaState
		if !ok {
			continue
		}
		if prev != 0 || player.PersonaState == 0 {
			continue
		}

		payload := map[string]any{
			"steam_id":      cfg.SteamID,
			"persona_state": player.PersonaState,
			"personaname":   player.PersonaName,
			"game_id":       player.GameID,
			"game_name":     player.GameExtraInfo,
			"timestamp":     time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
			payload["content"] = fmt.Sprintf("Steam: %s is online", player.PersonaName)
		}
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("steam poller trigger wf %d: %v", wf.ID, err)
		}
	}
}

func pollGameSales(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, lastSale map[int64]int, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "steam_game_sale")
	if err != nil {
		log.Printf("steam poller: list sale workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.SteamGameSaleConfigFromJSON(wf.TriggerConfig)
		if err != nil || cfg.AppID <= 0 {
			log.Printf("steam poller wf %d: bad config: %v", wf.ID, err)
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

		app, err := fetchAppDetails(ctx, cfg.AppID, cfg.Country)
		if err != nil {
			log.Printf("steam poller wf %d: app details: %v", wf.ID, err)
			continue
		}
		if app.PriceOverview == nil {
			continue
		}
		prev, ok := lastSale[wf.ID]
		lastSale[wf.ID] = app.PriceOverview.DiscountPercent
		if !ok {
			continue
		}
		if prev > 0 || app.PriceOverview.DiscountPercent <= 0 {
			continue
		}

		payload := map[string]any{
			"app_id":           cfg.AppID,
			"name":             app.Name,
			"discount_percent": app.PriceOverview.DiscountPercent,
			"initial_price":    app.PriceOverview.Initial,
			"final_price":      app.PriceOverview.Final,
			"currency":         app.PriceOverview.Currency,
			"timestamp":        time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
			payload["content"] = fmt.Sprintf("Steam sale: %s (-%d%%)", app.Name, app.PriceOverview.DiscountPercent)
		}
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("steam poller trigger sale wf %d: %v", wf.ID, err)
		}
	}
}

func pollPriceChanges(ctx context.Context, wfStore *workflows.Store, wfService *workflows.Service, lastPrice map[int64]int, lastCheck map[int64]time.Time) {
	wfs, err := wfStore.ListWorkflowsByTrigger(ctx, "steam_price_change")
	if err != nil {
		log.Printf("steam poller: list price workflows: %v", err)
		return
	}
	for _, wf := range wfs {
		if !wf.Enabled {
			continue
		}
		cfg, err := workflows.SteamPriceChangeConfigFromJSON(wf.TriggerConfig)
		if err != nil || cfg.AppID <= 0 {
			log.Printf("steam poller wf %d: bad config: %v", wf.ID, err)
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

		app, err := fetchAppDetails(ctx, cfg.AppID, cfg.Country)
		if err != nil {
			log.Printf("steam poller wf %d: app details: %v", wf.ID, err)
			continue
		}
		if app.PriceOverview == nil {
			continue
		}
		prev, ok := lastPrice[wf.ID]
		lastPrice[wf.ID] = app.PriceOverview.Final
		if !ok {
			continue
		}
		if prev == app.PriceOverview.Final {
			continue
		}

		payload := map[string]any{
			"app_id":           cfg.AppID,
			"name":             app.Name,
			"old_price":        prev,
			"new_price":        app.PriceOverview.Final,
			"currency":         app.PriceOverview.Currency,
			"discount_percent": app.PriceOverview.DiscountPercent,
			"timestamp":        time.Now().Format(time.RFC3339),
		}
		for k, v := range cfg.PayloadTemplate {
			payload[k] = v
		}
		if content, ok := payload["content"]; !ok || fmt.Sprint(content) == "" {
			payload["content"] = fmt.Sprintf("Steam price change: %s (%d -> %d %s)", app.Name, prev, app.PriceOverview.Final, app.PriceOverview.Currency)
		}
		if _, err := wfService.Trigger(workflows.WithUserID(ctx, wf.UserID), wf.ID, payload); err != nil {
			log.Printf("steam poller trigger price wf %d: %v", wf.ID, err)
		}
	}
}

type steamPlayerSummary struct {
	SteamID      string `json:"steamid"`
	PersonaName  string `json:"personaname"`
	PersonaState int    `json:"personastate"`
	GameID       string `json:"gameid"`
	GameExtraInfo string `json:"gameextrainfo"`
}

func fetchPlayerSummary(ctx context.Context, apiKey, steamID string) (*steamPlayerSummary, error) {
	u := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		url.QueryEscape(apiKey),
		url.QueryEscape(strings.TrimSpace(steamID)),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steam status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Response struct {
			Players []steamPlayerSummary `json:"players"`
		} `json:"response"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if len(payload.Response.Players) == 0 {
		return nil, fmt.Errorf("steam user not found")
	}
	return &payload.Response.Players[0], nil
}

type steamPriceOverview struct {
	Currency        string `json:"currency"`
	Initial         int    `json:"initial"`
	Final           int    `json:"final"`
	DiscountPercent int    `json:"discount_percent"`
}

type steamAppData struct {
	Name          string             `json:"name"`
	PriceOverview *steamPriceOverview `json:"price_overview"`
}

func fetchAppDetails(ctx context.Context, appID int, country string) (*steamAppData, error) {
	cc := strings.TrimSpace(strings.ToLower(country))
	if cc == "" {
		cc = "us"
	}
	u := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&cc=%s&l=en", appID, url.QueryEscape(cc))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steam store status %d: %s", resp.StatusCode, string(body))
	}
	var raw map[string]struct {
		Success bool         `json:"success"`
		Data    *steamAppData `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	key := strconv.Itoa(appID)
	entry, ok := raw[key]
	if !ok || !entry.Success || entry.Data == nil {
		return nil, fmt.Errorf("steam app not found")
	}
	return entry.Data, nil
}
