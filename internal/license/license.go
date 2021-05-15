package license

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/PuerkitoBio/goquery"
)

var (
	ErrModuleNotFound  = errors.New("module not found")
	ErrLicenseNotFound = errors.New("no license found")
)

func Resolve(module gomod.Module) ([]*SPDXLicense, error) {
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

func resolveForCoordinates(coordinates string) ([]*SPDXLicense, error) {
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

	licenses := make([]*SPDXLicense, 0)
	for _, licenseID := range strings.Split(licenseIDs, ",") {
		if license := getLicenseByID(strings.TrimSpace(licenseID)); license != nil {
			licenses = append(licenses, license)
		} else {
			log.Printf("the resolved license ID %s is not a valid SPDX license ID\n", licenseID)
		}
	}

	return licenses, nil
}
