package test

import (
	"archive/tar"
	"bytes"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	testResources "github.com/ngyewch/go-ssh-helper/test/resources"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	openSshServerRepoTag       = "lscr.io/linuxserver/openssh-server:9.3_p2-r0-ls127"
	customOpenSshServerRepoTag = "go-ssh-helper:latest"
)

type OpenSSHServerManager struct {
	client     *docker.Client
	containers []*docker.Container
	networks   []*docker.Network
}

func NewOpenSSHServerManager() (*OpenSSHServerManager, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	manager := &OpenSSHServerManager{
		client: client,
	}

	err = manager.PullImage(openSshServerRepoTag)
	if err != nil {
		return nil, err
	}

	dockerBuildContextFs, err := fs.Sub(testResources.DockerBuildContextFS, "dockerBuildContext")
	if err != nil {
		return nil, err
	}
	err = manager.BuildImage(customOpenSshServerRepoTag, dockerBuildContextFs)

	return manager, nil
}

func (manager *OpenSSHServerManager) Close() error {
	for _, container := range manager.containers {
		err := manager.client.StopContainer(container.ID, 15)
		if err != nil {
			return err
		}

		err = manager.client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
		if err != nil {
			return err
		}
	}

	for _, network := range manager.networks {
		err := manager.client.RemoveNetwork(network.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *OpenSSHServerManager) PullImage(repoTag string) error {
	return manager.client.PullImage(docker.PullImageOptions{
		Repository:   repoTag,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
}

func (manager *OpenSSHServerManager) BuildImage(imageName string, dockerBuildContextFs fs.FS) error {
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

	err = manager.client.BuildImage(docker.BuildImageOptions{
		Name:         imageName,
		InputStream:  inputBuf,
		OutputStream: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func (manager *OpenSSHServerManager) CreateNetwork(networkId string) (*docker.Network, error) {
	network, err := manager.client.CreateNetwork(docker.CreateNetworkOptions{
		Name: networkId,
	})
	if err != nil {
		return nil, err
	}

	manager.networks = append(manager.networks, network)

	return network, nil
}

func (manager *OpenSSHServerManager) StartSshHost(hostname string, username string, publicKey string, hostPort int) (*docker.Container, error) {
	createContainerOptions := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        customOpenSshServerRepoTag,
			AttachStdout: true,
			AttachStderr: true,
			PortSpecs: []string{
				"2222/tcp",
			},
			Env: []string{
				fmt.Sprintf("PUBLIC_KEY=%s", publicKey),
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
	container, err := manager.client.CreateContainer(createContainerOptions)
	if err != nil {
		return nil, err
	}

	err = manager.client.StartContainer(container.ID, &docker.HostConfig{
		AutoRemove: true,
	})
	if err != nil {
		return nil, err
	}

	manager.containers = append(manager.containers, container)

	return container, nil
}

func (manager *OpenSSHServerManager) ConnectNetworks(container *docker.Container, networks ...*docker.Network) error {
	for _, network := range networks {
		err := manager.client.ConnectNetwork(network.ID, docker.NetworkConnectionOptions{
			Container: container.ID,
		})
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
			defer func(f fs.File) {
				_ = f.Close()
			}(f)
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
