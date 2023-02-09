package identity

import (
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/crypto"
	poktCrypto "github.com/pokt-network/pocket/shared/crypto"
)

func NewLibP2PPrivateKey(hexString string) (crypto.PrivKey, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, poktCrypto.ErrCreatePrivateKey(err)
	}

	privateKey, err := crypto.PrivKeyUnmarshallers[crypto.Ed25519](bz)
	if err != nil {
		return nil, poktCrypto.ErrCreatePublicKey(err)
	}

	return privateKey, nil
}
