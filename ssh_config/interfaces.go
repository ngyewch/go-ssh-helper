package ssh_config

type SshConfig interface {
	Get(alias string, key string) string
	GetAll(alias string, key string) []string
}
