package test

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"gotest.tools/v3/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func TestHostA(t *testing.T) {
	doCheckHostname(t, "test-hostA", "hostA")
}

func TestHostB(t *testing.T) {
	doCheckHostname(t, "test-hostB", "hostB")
}

func TestHostC(t *testing.T) {
	doCheckHostname(t, "test-hostC", "hostC")
}

func TestHostC2(t *testing.T) {
	doCheckHostname(t, "test-hostC-2", "hostC")
}

func doCheckHostname(t *testing.T, alias string, expectedHostname string) {
	sshClient, err := testEnv.sshClientFactory.CreateForAlias(alias)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func(sshClient *ssh.Client) {
		_ = sshClient.Close()
	}(sshClient)

	outBytes, _, err := sessionRun(sshClient, "hostname")
	if err != nil {
		t.Fatal(err)
		return
	}

	output := strings.TrimSpace(string(outBytes))
	assert.Equal(t, expectedHostname, output, "expected %s, actual %s", expectedHostname, output)
}

func sessionRun(sshClient *ssh.Client, cmd string) ([]byte, []byte, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, nil, err
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	outBuffer := bytes.NewBuffer(nil)
	errBuffer := bytes.NewBuffer(nil)

	session.Stdout = io.MultiWriter(os.Stdout, outBuffer)
	session.Stderr = io.MultiWriter(os.Stderr, errBuffer)

	err = session.Run(cmd)
	if err != nil {
		return nil, nil, err
	}

	return outBuffer.Bytes(), errBuffer.Bytes(), nil
}
