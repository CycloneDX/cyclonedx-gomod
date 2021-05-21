package license

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/PuerkitoBio/goquery"
)

var (
	ErrModuleNotFound  = errors.New("module not found")
	ErrLicenseNotFound = errors.New("no license found")
	ErrLocalModule     = errors.New("license resolution isn't supported for local modules")
)

func Resolve(module gomod.Module) ([]cdx.License, error) {
	if module.Local {
		return nil, ErrLocalModule
	}

	licenses, err := resolveForCoordinates(module.Coordinates())
	if err != nil {
		// The specific version of the module may not be present
		// in the module proxy yet. Retry with just he module path
		if errors.Is(err, ErrModuleNotFound) {
			return resolveForCoordinates(module.Path)
		}
		return nil, err
	}
	return licenses, nil
}

func resolveForCoordinates(coordinates string) ([]cdx.License, error) {
	req, err := http.NewRequest(http.MethodGet, "https://pkg.go.dev/"+coordinates, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.Name, version.Version))

	query := req.URL.Query()
	query.Add("tab", "licenses")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		return nil, ErrModuleNotFound
	default:
		return nil, fmt.Errorf("unexpected response status: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	sel := doc.Find("div.Container section.License h2").First()
	if len(sel.Nodes) == 0 {
		return nil, ErrLicenseNotFound
	}

	licenseIDs := strings.TrimSpace(sel.Text())

	licenses := make([]cdx.License, 0)
	for _, licenseID := range strings.Split(licenseIDs, ",") {
		licenseID = strings.TrimSpace(licenseID)
		if spdxLicense := getLicenseByID(licenseID); spdxLicense != nil {
			licenses = append(licenses, cdx.License{
				ID:  spdxLicense.ID,
				URL: spdxLicense.Reference,
			})
		} else {
			licenses = append(licenses, cdx.License{
				Name: licenseID,
			})
		}
	}

	return licenses, nil
}
