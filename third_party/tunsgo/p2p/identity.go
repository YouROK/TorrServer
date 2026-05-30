package p2p

import (
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func LoadOrCreateIdentity() (crypto.PrivKey, error) {
	dir := filepath.Dir(os.Args[0])
	filename := filepath.Join(dir, "node.key")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, err
		}
		data, _ := crypto.MarshalPrivateKey(priv)
		os.WriteFile(filename, data, 0600)
		return priv, nil
	}
	data, _ := os.ReadFile(filename)
	return crypto.UnmarshalPrivateKey(data)
}
