package config

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Identity struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}

func (i *Identity) DecodePrivateKey(passphrase string) (crypto.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(i.PrivKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(pkb)
}

// CreateIdentity initializes a new identity.
func CreateIdentity() (Identity, error) {
	ident := Identity{}

	var err error
	var sk crypto.PrivKey
	var pk crypto.PubKey

	sk, pk, err = crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return ident, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()
	return ident, nil
}
