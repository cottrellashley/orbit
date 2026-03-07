package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/cottrellashley/orbit/internal/role"
	"github.com/spf13/cobra"
)

// ListCmd creates the `orbit list` command.
func ListCmd() *cobra.Command {
	var tag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			for _, r := range cfg.Roles {
				// Filter by tag if specified
				if tag != "" && !hasTag(r, tag) {
					continue
				}

				status := "ok"
				if _, err := os.Stat(r.Path); os.IsNotExist(err) {
					status = "missing"
				}

				extra := ""
				if r.Type == role.Workspace {
					count := countChildren(r.Path)
					extra = fmt.Sprintf(" (%d projects)", count)
				}

				tags := ""
				if len(r.Tags) > 0 {
					tags = " [" + strings.Join(r.Tags, ", ") + "]"
				}

				fmt.Printf("%-20s %-12s %-8s %s%s%s\n", r.Name, r.Type, status, r.Path, extra, tags)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "filter roles by tag")

	return cmd
}

func hasTag(r *role.Role, tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func countChildren(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			count++
		}
	}
	return count
}
