package ssh_config

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"testing"
)

func Test1(t *testing.T) {
	sshClient, err := NewSshClientForAlias(os.Getenv("HOST"))
	if err != nil {
		t.Fatal(err)
	}
	defer func(sshClient *ssh.Client) {
		_ = sshClient.Close()
	}(sshClient)

	session, err := sshClient.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run("hostname")
	if err != nil {
		t.Fatal(err)
	}

	hostname := strings.TrimSpace(b.String())
	t.Logf("hostname: %s\n", hostname)
}
