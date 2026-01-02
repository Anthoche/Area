package youtube

import (
	"context"
	"testing"

	"area/src/integrations/youtube"
)

func TestStartYouTubePoller_NoDeps(t *testing.T) {
	youtube.StartYouTubePoller(context.Background(), nil, nil)
}
