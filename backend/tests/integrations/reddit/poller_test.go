package reddit

import (
	"context"
	"testing"

	"area/src/integrations/reddit"
)

func TestStartRedditPoller_NoDeps(t *testing.T) {
	reddit.StartRedditPoller(context.Background(), nil, nil)
}
