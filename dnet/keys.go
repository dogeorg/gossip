package dnet

import (
	"crypto/ed25519"
	"crypto/rand"
)

type PrivKey = ed25519.PrivateKey // [32]privkey then [32]pubkey
type PubKey = ed25519.PublicKey   // [32]pubkey

type KeyPair struct {
	Priv ed25519.PrivateKey
	Pub  ed25519.PublicKey
}

// Make a KeyPair from 32-byte Entropy Seed.
// e.g. from BIP39 mnemonic phrase or cryptographically secure random data.
func KeyPairFromSeed(seed []byte) KeyPair {
	priv := ed25519.NewKeyFromSeed(seed)
	return KeyPairFromPrivKey(priv)
}

// Make a KeyPair from a 64-byte ed25519 PrivateKey.
func KeyPairFromPrivKey(priv PrivKey) KeyPair {
	pub := make([]byte, ed25519.PublicKeySize)
	copy(pub, priv[32:]) // make PubKey safe to pass around
	return KeyPair{Priv: priv, Pub: ed25519.PublicKey(pub)}
}

// Generate a new Private Key using "crypto/rand" entropy source.
// Can return an error if there is insufficient entropy available.
func GenerateKeyPair() (KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return KeyPair{Priv: priv, Pub: pub}, err
}
