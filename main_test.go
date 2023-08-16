package ssh_helper

import (
	"golang.org/x/crypto/ssh"
	"testing"
)

func Test1(t *testing.T) {
	sshClient, err := SshClientFromSshConfig("test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(sshClient *ssh.Client) {
		_ = sshClient.Close()
	}(sshClient)
}
