package test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"github.com/ngyewch/go-ssh-helper"
	"github.com/trzsz/ssh_config"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type TestEnv struct {
	tmpDir               string
	openSshServerManager *OpenSSHServerManager
	sshClientFactory     *ssh_helper.SSHClientFactory
}

func NewTestEnv() (*TestEnv, error) {
	test := &TestEnv{}

	{
		tmpDir, err := os.MkdirTemp("", "ssh_config-*")
		if err != nil {
			return nil, err
		}
		test.tmpDir = tmpDir

		sshConfigPath := filepath.Join(tmpDir, "test")
		err = func() error {
			testResourceFs := os.DirFS(filepath.Join("resources", "ssh_config"))
			r, err := testResourceFs.Open("test")
			if err != nil {
				return err
			}
			defer func(r fs.File) {
				_ = r.Close()
			}(r)

			w, err := os.Create(sshConfigPath)
			if err != nil {
				return err
			}
			defer func(w *os.File) {
				_ = w.Close()
			}(w)

			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}

			return nil
		}()

		userSettings := &ssh_config.UserSettings{}
		userSettings.ConfigFinder(func() string {
			return sshConfigPath
		})

		test.sshClientFactory = ssh_helper.NewSSHClientFactory(userSettings)
	}
	{
		openSshServerManager, err := NewOpenSSHServerManager()
		if err != nil {
			return nil, err
		}
		test.openSshServerManager = openSshServerManager
	}

	return test, nil
}

func (test *TestEnv) Setup() error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	publicKey, err := func() (string, error) {
		publicKeyFile, err := homedir.Expand("~/.ssh/id_rsa.pub")
		if err != nil {
			return "", err
		}
		publicKeyBytes, err := os.ReadFile(publicKeyFile)
		if err != nil {
			return "", err
		}
		return string(publicKeyBytes), nil
	}()

	network1, err := test.openSshServerManager.CreateNetwork(fmt.Sprintf("network1_%s", id))
	if err != nil {
		return err
	}

	network2, err := test.openSshServerManager.CreateNetwork(fmt.Sprintf("network2_%s", id))
	if err != nil {
		return err
	}

	containerA, err := test.openSshServerManager.StartSshHost("hostA", "userA", publicKey, 2222)
	if err != nil {
		return err
	}

	err = test.openSshServerManager.ConnectNetworks(containerA, network1)
	if err != nil {
		return err
	}

	containerB, err := test.openSshServerManager.StartSshHost("hostB", "userB", publicKey, 0)
	if err != nil {
		return err
	}

	err = test.openSshServerManager.ConnectNetworks(containerB, network1, network2)
	if err != nil {
		return err
	}

	containerC, err := test.openSshServerManager.StartSshHost("hostC", "userC", publicKey, 0)
	if err != nil {
		return err
	}

	err = test.openSshServerManager.ConnectNetworks(containerC, network2)
	if err != nil {
		return err
	}

	return nil
}

func (test *TestEnv) Close() error {
	if test.tmpDir != "" {
		err := os.RemoveAll(test.tmpDir)
		if err != nil {
			return err
		}
	}
	if test.openSshServerManager != nil {
		err := test.openSshServerManager.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
