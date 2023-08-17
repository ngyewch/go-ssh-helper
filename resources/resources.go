package resources

import "embed"

var (
	//go:embed dockerBuildContext
	DockerBuildContextFS embed.FS

	//go:embed ssh_config
	SshConfigFS embed.FS
)
