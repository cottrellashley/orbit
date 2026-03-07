package commands

import (
	"fmt"

	"github.com/cottrellashley/orbit/internal/config"
	"github.com/spf13/cobra"
)

// loadConfig reads the config from the path specified by --config flag,
// or the default location.
func loadConfig(cmd *cobra.Command) (*config.Config, string, error) {
	path, _ := cmd.Root().Flags().GetString("config")
	if path == "" {
		path = config.DefaultPath()
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, path, fmt.Errorf("cannot load config: %w", err)
	}
	return cfg, path, nil
}

// saveConfig writes the config back to the given path.
func saveConfig(path string, cfg *config.Config) error {
	return config.Save(path, cfg)
}
