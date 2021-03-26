package gomod

import (
	"fmt"
	"os"

	"golang.org/x/mod/sumdb/dirhash"
)

func (m Module) Hash() (string, error) {
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return "", fmt.Errorf("module dir %s does not exist", m.Dir)
	}

	h1, err := dirhash.HashDir(m.Dir, m.Coordinates(), dirhash.Hash1)
	if err != nil {
		return "", err
	}

	return h1, nil
}
