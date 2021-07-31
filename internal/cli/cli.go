package cli

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type OutputOptions struct {
	FilePath string
	UseJSON  bool
}

func (o *OutputOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.UseJSON, "json", false, "Output in JSON")
	fs.StringVar(&o.FilePath, "out", "-", "Output file path")
}

type SBOMOptions struct {
	NoSerialNumber  bool
	NoVersionPrefix bool
	Reproducible    bool
	SerialNumber    string
}

func (s *SBOMOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.NoSerialNumber, "noserial", false, "Omit serial number")
	fs.BoolVar(&s.NoVersionPrefix, "novprefix", false, "Omit \"v\" prefix from versions")
	fs.BoolVar(&s.Reproducible, "reproducible", false, "Make the SBOM reproducible by omitting dynamic content")
	fs.StringVar(&s.SerialNumber, "serial", "", "Serial number")
}

func NewRootCmd() *ffcli.Command {
	return &ffcli.Command{
		Name: "cyclonedx-gomod",
		Subcommands: []*ffcli.Command{
			newModCmd(),
			newVersionCmd(),
		},
		Exec: func(_ context.Context, _ []string) error {
			return flag.ErrHelp
		},
	}
}
