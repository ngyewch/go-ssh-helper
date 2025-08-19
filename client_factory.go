package ssh_helper

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/trzsz/ssh_config"
	"golang.org/x/crypto/ssh"
)

var (
	defaultSshClientFactory = NewSSHClientFactory(&ssh_config.UserSettings{})
)

// SSHClientFactory is a factory for creating ssh.Client.
type SSHClientFactory struct {
	userSettings *ssh_config.UserSettings
}

// NewSSHClientFactory instantiates a new SSHClientFactory.
func NewSSHClientFactory(userSettings *ssh_config.UserSettings) *SSHClientFactory {
	return &SSHClientFactory{
		userSettings: userSettings,
	}
}

// DefaultSSHClientFactory returns the default SSHClientFactory
func DefaultSSHClientFactory() *SSHClientFactory {
	return defaultSshClientFactory
}

// CreateForAlias creates a new ssh.Client for the specified alias.
func (factory *SSHClientFactory) CreateForAlias(alias string) (*ssh.Client, error) {
	return factory.doCreateForAlias(alias, nil)
}

func (factory *SSHClientFactory) doCreateForAlias(alias string, baseClient *ssh.Client) (*ssh.Client, error) {
	aliasHost := alias
	aliasUser := ""
	if strings.Contains(alias, "@") {
		p := strings.LastIndex(alias, "@")
		aliasUser = alias[:p]
		aliasHost = alias[p+1:]
	}
	hostname := factory.userSettings.Get(aliasHost, "Hostname")
	if hostname == "" {
		return nil, fmt.Errorf("alias does not exist")
	}

	port := factory.userSettings.Get(aliasHost, "Port")

	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshConfig.User = factory.userSettings.Get(aliasHost, "User")
	if aliasUser != "" {
		sshConfig.User = aliasUser
	}

	sTimeoutInSeconds := factory.userSettings.Get(aliasHost, "ConnectTimeout")
	if sTimeoutInSeconds != "" {
		timeoutInSeconds, err := strconv.Atoi(sTimeoutInSeconds)
		if err != nil {
			return nil, err
		}
		sshConfig.Timeout = time.Duration(timeoutInSeconds) * time.Second
	}

	identityFiles := factory.userSettings.GetAll(aliasHost, "IdentityFile")
	for _, identityFile := range identityFiles {
		identityFile, err := homedir.Expand(identityFile)
		if err != nil {
			return nil, err
		}
		signer, err := LoadSignerFromFile(identityFile)
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	network := "tcp"
	switch factory.userSettings.Get(aliasHost, "AddressFamily") {
	case "any":
		network = "tcp"
	case "inet":
		network = "tcp4"
	case "inet6":
		network = "tcp6"
	}

	var proxyJumps []string
	if factory.userSettings.Get(aliasHost, "ProxyJump") != "" {
		proxyJumps = strings.Split(factory.userSettings.Get(aliasHost, "ProxyJump"), ",")
		for i := 0; i < len(proxyJumps); i++ {
			proxyJumps[i] = strings.TrimSpace(proxyJumps[i])
		}
	}

	proxyClient := baseClient
	var proxyClients []*ssh.Client
	for _, proxyJump := range proxyJumps {
		if proxyJump != "" {
			newProxyClient, err := factory.doCreateForAlias(proxyJump, proxyClient)
			if err != nil {
				for _, proxyClient1 := range proxyClients {
					_ = proxyClient1.Close()
				}
				return nil, err
			}
			proxyClients = append(proxyClients, newProxyClient)
			proxyClient = newProxyClient
		}
	}

	if proxyClient != nil {
		conn, err := proxyClient.Dial(network, net.JoinHostPort(hostname, port))
		if err != nil {
			return nil, err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(hostname, port), sshConfig)
		if err != nil {
			return nil, err
		}

		client := ssh.NewClient(ncc, chans, reqs)
		go func() {
			_ = client.Wait()
			for _, proxyClient1 := range proxyClients {
				_ = proxyClient1.Close()
			}
		}()
		return client, nil
	} else {
		client, err := ssh.Dial(network, net.JoinHostPort(hostname, port), sshConfig)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
