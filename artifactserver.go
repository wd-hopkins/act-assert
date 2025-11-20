package act_assert

type ArtifactServerConfig struct {
	// Host Defines the address to which the artifact server binds.
	Host string
	// Port Defines the port where the artifact server listens.
	Port int
	// Path Defines the path where the artifact server stores uploads and retrieves downloads from. If not specified the artifact server will not start.
	Path string
	// Cleanup Indicates whether to clean up the artifact storage path after use.
	Cleanup bool
}
