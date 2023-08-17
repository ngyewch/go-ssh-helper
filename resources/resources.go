package resources

import "embed"

var (
	//go:embed test/dockerBuildContext
	DockerBuildContextFS embed.FS

	//go:embed test/ssh_config
	SshConfigFS embed.FS
)
