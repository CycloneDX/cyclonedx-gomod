package bin

import (
	"errors"
	"flag"
	"fmt"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
)

type BinOptions struct {
	options.OutputOptions
	options.SBOMOptions

	BinaryPath string
	Version    string
}

func (b *BinOptions) RegisterFlags(fs *flag.FlagSet) {
	b.OutputOptions.RegisterFlags(fs)
	b.SBOMOptions.RegisterFlags(fs)

	fs.StringVar(&b.Version, "version", "", "Version of the main component")
}

func (b BinOptions) Validate() error {
	errs := make([]error, 0)

	if err := b.OutputOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := b.SBOMOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	if !util.FileExists(b.BinaryPath) {
		errs = append(errs, fmt.Errorf("binary at %s does not exist", b.BinaryPath))
	}

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}
