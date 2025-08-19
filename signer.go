package ssh_helper

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

// LoadSignerFromFile instantiates an ssh.Signer from a file containing a private key.
func LoadSignerFromFile(path string) (ssh.Signer, error) {
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}
	privateKeyBytes, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	return signer, nil
}
