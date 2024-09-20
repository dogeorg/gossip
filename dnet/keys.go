package dnet

import (
	"github.com/dogeorg/doge"
)

type PrivKey = doge.ECPrivKey
type PubKey = doge.ECPubKeySchnorr

type KeyPair struct {
	Priv PrivKey
	Pub  PubKey
}

// Make a KeyPair from a 32-byte private key.
func KeyPairFromPrivKey(priv PrivKey) KeyPair {
	pub := doge.ECPubKeyFromECPrivKey(priv)
	// Schnorr pubkeys are X-only (ignore the Y-odd/even byte)
	pubX := (*[32]byte)(pub[1:33])
	return KeyPair{Priv: priv, Pub: pubX}
}

// Generate a new Private Key using "crypto/rand" entropy source.
// Can return an error if there is insufficient entropy available.
func GenerateKeyPair() (KeyPair, error) {
	priv, err := doge.GenerateECPrivKey()
	if err != nil {
		return KeyPair{}, nil
	}
	return KeyPairFromPrivKey(priv), nil
}
