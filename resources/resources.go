package resources

import "embed"

var (
	//go:embed dockerBuildContext
	DockerBuildContextFS embed.FS
)
