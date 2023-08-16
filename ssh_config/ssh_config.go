package ssh_config

import (
	"fmt"
	"github.com/kevinburke/ssh_config"
	"github.com/mitchellh/go-homedir"
	"github.com/ngyewch/go-ssh-helper/common"
	"golang.org/x/crypto/ssh"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	userSettings = ssh_config.UserSettings{}
)

func NewSshClientForAlias(alias string) (*ssh.Client, error) {
	return doNewSshClientForAlias(alias)
}

func doNewSshClientForAlias(alias string) (*ssh.Client, error) {
	fmt.Printf("creating %s ...\n", alias)

	aliasHost := alias
	aliasUser := ""
	if strings.Contains(alias, "@") {
		p := strings.LastIndex(alias, "@")
		aliasUser = alias[:p]
		aliasHost = alias[p+1:]
	}
	hostname := userSettings.Get(aliasHost, "Hostname")
	if hostname == "" {
		return nil, fmt.Errorf("alias does not exist")
	}

	var proxyClient *ssh.Client
	proxyJump := userSettings.Get(aliasHost, "ProxyJump")
	if proxyJump != "" {
		newProxyClient, err := doNewSshClientForAlias(proxyJump)
		if err != nil {
			return nil, err
		}
		proxyClient = newProxyClient
	}

	port := userSettings.Get(aliasHost, "Port")

	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshConfig.User = userSettings.Get(aliasHost, "User")
	if aliasUser != "" {
		sshConfig.User = aliasUser
	}

	sTimeoutInSeconds := userSettings.Get(aliasHost, "ConnectTimeout")
	if sTimeoutInSeconds != "" {
		timeoutInSeconds, err := strconv.Atoi(sTimeoutInSeconds)
		if err != nil {
			return nil, err
		}
		sshConfig.Timeout = time.Duration(timeoutInSeconds) * time.Second
	}

	identityFiles := userSettings.GetAll(aliasHost, "IdentityFile")
	for _, identityFile := range identityFiles {
		identityFile, err := homedir.Expand(identityFile)
		if err != nil {
			return nil, err
		}
		signer, err := common.LoadSignerFromFile(identityFile)
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	network := "tcp"
	switch userSettings.Get(aliasHost, "AddressFamily") {
	case "any":
		network = "tcp"
	case "inet":
		network = "tcp4"
	case "inet6":
		network = "tcp6"
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
		return client, nil
	} else {
		client, err := ssh.Dial(network, net.JoinHostPort(hostname, port), sshConfig)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
