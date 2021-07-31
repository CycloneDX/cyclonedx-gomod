package cli

import (
	"context"
	"fmt"

	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func newVersionCmd() *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortHelp:  "Show version information",
		ShortUsage: "cyclonedx-gomod version",
		Exec: func(_ context.Context, _ []string) error {
			fmt.Println(version.Version)
			return nil
		},
	}
}
