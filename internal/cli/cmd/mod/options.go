package mod

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

// ModOptions provides options for the `mod` command.
type ModOptions struct {
	options.OutputOptions
	options.SBOMOptions

	ComponentType   string
	ModuleDir       string
	IncludeTest     bool
	ResolveLicenses bool
}

func (m *ModOptions) RegisterFlags(fs *flag.FlagSet) {
	m.OutputOptions.RegisterFlags(fs)
	m.SBOMOptions.RegisterFlags(fs)

	fs.StringVar(&m.ComponentType, "type", "application", "Type of the main component")
	fs.BoolVar(&m.IncludeTest, "test", false, "Include test dependencies")
	fs.BoolVar(&m.ResolveLicenses, "licenses", false, "Resolve module licenses")
}

var allowedComponentTypes = []cdx.ComponentType{
	cdx.ComponentTypeApplication,
	cdx.ComponentTypeFirmware,
	cdx.ComponentTypeFramework,
	cdx.ComponentTypeLibrary,
}

func (m ModOptions) Validate() error {
	errs := make([]error, 0)

	if err := m.OutputOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := m.SBOMOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	isAllowedComponentType := false
	for i := range allowedComponentTypes {
		if allowedComponentTypes[i] == cdx.ComponentType(m.ComponentType) {
			isAllowedComponentType = true
			break
		}
	}
	if !isAllowedComponentType {
		allowed := make([]string, len(allowedComponentTypes))
		for i := range allowedComponentTypes {
			allowed[i] = string(allowedComponentTypes[i])
		}

		errs = append(errs, fmt.Errorf("invalid component type: \"%s\" (allowed: %s)", m.ComponentType, strings.Join(allowed, ",")))
	}

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}
