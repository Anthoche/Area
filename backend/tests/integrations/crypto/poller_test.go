package crypto

import (
	"context"
	"testing"

	"area/src/integrations/crypto"
)

func TestStartCryptoPoller_NoDeps(t *testing.T) {
	crypto.StartCryptoPoller(context.Background(), nil, nil)
}
