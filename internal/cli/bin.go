package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type BinOptions struct {
	OutputOptions
	SBOMOptions

	BinaryPath string
}

func (b *BinOptions) RegisterFlags(fs *flag.FlagSet) {
	b.OutputOptions.RegisterFlags(fs)
	b.SBOMOptions.RegisterFlags(fs)
}

func (b BinOptions) Validate() error {
	if !util.FileExists(b.BinaryPath) {
		return &OptionsValidationError{Errors: []error{fmt.Errorf("binary at %s does not exist", b.BinaryPath)}}
	}

	return nil
}

func newBinCmd() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod bin", flag.ExitOnError)

	var options BinOptions
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "bin",
		ShortHelp:  "Generate SBOM for a binary",
		ShortUsage: "cyclonedx-gomod bin [FLAGS...] PATH",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("no binary path provided")
			}

			options.BinaryPath = args[0]
			return execBinCmd(options)
		},
	}
}

func execBinCmd(options BinOptions) error {
	if err := options.Validate(); err != nil {
		return err
	}

	return nil
}
