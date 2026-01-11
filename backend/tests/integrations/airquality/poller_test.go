package airquality

import (
	"context"
	"testing"

	"area/src/integrations/airquality"
)

func TestStartAirQualityPoller_NoDeps(t *testing.T) {
	airquality.StartAirQualityPoller(context.Background(), nil, nil)
}
