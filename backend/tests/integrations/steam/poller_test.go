package steam

import (
	"context"
	"testing"

	"area/src/integrations/steam"
)

func TestStartSteamPoller_NoDeps(t *testing.T) {
	steam.StartSteamPoller(context.Background(), nil, nil)
}
