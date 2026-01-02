package weather

import (
	"context"
	"testing"

	weath "area/src/integrations/weather"
)

func TestStartWeatherPoller_NoDeps(t *testing.T) {
	weath.StartWeatherPoller(context.Background(), nil, nil)
}
