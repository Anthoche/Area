package nasa

import (
	"context"
	"testing"

	"area/src/integrations/nasa"
)

func TestStartNasaPoller_NoDeps(t *testing.T) {
	nasa.StartNasaPoller(context.Background(), nil, nil)
}
