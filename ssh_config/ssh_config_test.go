package ssh_config

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"github.com/ngyewch/go-ssh-helper/resources"
	"golang.org/x/crypto/ssh"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"
)

type TestEnv struct {
	client     *docker.Client
	uuid       string
	publicKey  string
	containers []*docker.Container
	networks   []*docker.Network
}

const (
	openSshServerRepoTag       = "lscr.io/linuxserver/openssh-server:9.3_p2-r0-ls127"
	customOpenSshServerRepoTag = "go-ssh-helper:latest"
)

var (
	testEnv = &TestEnv{}
)

func TestMain(m *testing.M) {
	testEnv0, err := NewTestEnv()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	testEnv = testEnv0

	err = testEnv.Setup()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	// wait for containers to start
	time.Sleep(10 * time.Second)

	code := m.Run()

	err = testEnv.Close()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func Test1(t *testing.T) {
	sshClient, err := NewSshClientForAlias(os.Getenv("HOST"))
	if err != nil {
		t.Fatal(err)
	}
	defer func(sshClient *ssh.Client) {
		_ = sshClient.Close()
	}(sshClient)

	err = sessionRun(sshClient, "hostname")
	if err != nil {
		t.Fatal(err)
	}

	/*
		err = sessionRun(sshClient, "ping -c 5 hostC")
		if err != nil {
			t.Fatal(err)
		}

		err = sessionRun(sshClient, "nc -v -v -v -N hostC 2222")
		if err != nil {
			t.Fatal(err)
		}
	*/
}

func sessionRun(sshClient *ssh.Client, cmd string) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run(cmd)
	if err != nil {
		return err
	}

	return nil
}

func NewTestEnv() (*TestEnv, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	publicKeyFile, err := homedir.Expand("~/.ssh/id_rsa.pub")
	if err != nil {
		return nil, err
	}
	publicKeyBytes, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, err
	}
	publicKey := string(publicKeyBytes)

	test := &TestEnv{
		client:    client,
		uuid:      id.String(),
		publicKey: publicKey,
	}

	err = test.pullImage(openSshServerRepoTag)
	if err != nil {
		return nil, err
	}

	dockerBuildContextFs, err := fs.Sub(resources.DockerBuildContextFS, "dockerBuildContext")
	if err != nil {
		return nil, err
	}
	err = test.buildImage(customOpenSshServerRepoTag, dockerBuildContextFs)

	return test, nil
}

func (test *TestEnv) pullImage(repoTag string) error {
	return test.client.PullImage(docker.PullImageOptions{
		Repository:   repoTag,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
}

func (test *TestEnv) buildImage(imageName string, dockerBuildContextFs fs.FS) error {
	inputBuf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputBuf)

	err := createTar(tr, dockerBuildContextFs, nil)
	if err != nil {
		return err
	}

	err = tr.Close()
	if err != nil {
		return err
	}

	err = test.client.BuildImage(docker.BuildImageOptions{
		Name:         imageName,
		InputStream:  inputBuf,
		OutputStream: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func (test *TestEnv) createNetwork(networkId string) (*docker.Network, error) {
	network, err := test.client.CreateNetwork(docker.CreateNetworkOptions{
		Name: networkId,
	})
	if err != nil {
		return nil, err
	}

	test.networks = append(test.networks, network)

	return network, nil
}

func (test *TestEnv) startSshHost(hostname string, username string, hostPort int) (*docker.Container, error) {
	createContainerOptions := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        customOpenSshServerRepoTag,
			AttachStdout: true,
			AttachStderr: true,
			PortSpecs: []string{
				"2222/tcp",
			},
			Env: []string{
				fmt.Sprintf("PUBLIC_KEY=%s", test.publicKey),
				fmt.Sprintf("USER_NAME=%s", username),
			},
			Hostname: hostname,
		},
	}
	if hostPort > 0 {
		createContainerOptions.Config.PortSpecs = []string{
			"2222/tcp",
		}
		createContainerOptions.HostConfig = &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"2222/tcp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: fmt.Sprintf("%d/tcp", hostPort),
					},
				},
			},
		}
	}
	container, err := test.client.CreateContainer(createContainerOptions)
	if err != nil {
		return nil, err
	}

	err = test.client.StartContainer(container.ID, &docker.HostConfig{
		AutoRemove: true,
	})
	if err != nil {
		return nil, err
	}

	test.containers = append(test.containers, container)

	return container, nil
}

func (test *TestEnv) Setup() error {
	network1, err := test.createNetwork(fmt.Sprintf("network1_%s", test.uuid))
	if err != nil {
		return err
	}

	network2, err := test.createNetwork(fmt.Sprintf("network2_%s", test.uuid))
	if err != nil {
		return err
	}

	containerA, err := test.startSshHost("hostA", "userA", 2222)
	if err != nil {
		return err
	}

	err = test.client.ConnectNetwork(network1.ID, docker.NetworkConnectionOptions{
		Container: containerA.ID,
	})
	if err != nil {
		return err
	}

	containerB, err := test.startSshHost("hostB", "userB", 0)
	if err != nil {
		return err
	}

	err = test.client.ConnectNetwork(network1.ID, docker.NetworkConnectionOptions{
		Container: containerB.ID,
	})
	if err != nil {
		return err
	}

	err = test.client.ConnectNetwork(network2.ID, docker.NetworkConnectionOptions{
		Container: containerB.ID,
	})
	if err != nil {
		return err
	}

	containerC, err := test.startSshHost("hostC", "userC", 0)
	if err != nil {
		return err
	}

	err = test.client.ConnectNetwork(network2.ID, docker.NetworkConnectionOptions{
		Container: containerC.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (test *TestEnv) Close() error {
	for _, container := range test.containers {
		err := test.client.StopContainer(container.ID, 15)
		if err != nil {
			return err
		}

		err = test.client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
		if err != nil {
			return err
		}
	}

	for _, network := range test.networks {
		err := test.client.RemoveNetwork(network.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTar(tr *tar.Writer, filesystem fs.FS, vars any) error {
	return fs.WalkDir(filesystem, ".", func(path string, entry fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		fi, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			t := time.Now()
			err = tr.WriteHeader(&tar.Header{
				Name:       path,
				Size:       0,
				Mode:       int64(fi.Mode()),
				ModTime:    t,
				AccessTime: t,
				ChangeTime: t,
			})
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(path, ".tmpl") {
			tmpl, err := template.New(filepath.Base(path)).ParseFS(filesystem, path)
			if err != nil {
				return err
			}
			outputBuf := bytes.NewBuffer(nil)
			err = tmpl.Execute(outputBuf, vars)
			if err != nil {
				return err
			}
			actualPath := path[0 : len(path)-5]
			contentBytes := outputBuf.Bytes()
			t := time.Now()
			err = tr.WriteHeader(&tar.Header{
				Name:       actualPath,
				Size:       int64(len(contentBytes)),
				Mode:       int64(fi.Mode()),
				ModTime:    t,
				AccessTime: t,
				ChangeTime: t,
			})
			if err != nil {
				return err
			}
			_, err = tr.Write(contentBytes)
			if err != nil {
				return err
			}
		} else {
			f, err := filesystem.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			contentBytes, err := io.ReadAll(f)
			t := time.Now()
			err = tr.WriteHeader(&tar.Header{
				Name:       path,
				Size:       int64(len(contentBytes)),
				Mode:       int64(fi.Mode()),
				ModTime:    t,
				AccessTime: t,
				ChangeTime: t,
			})
			if err != nil {
				return err
			}
			_, err = tr.Write(contentBytes)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
