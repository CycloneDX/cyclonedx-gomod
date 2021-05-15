package license

import (
	"errors"
	"fmt"
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

func Resolve(module gomod.Module) (string, error) {
	// TODO: Check for local Path

	req, err := http.NewRequest(http.MethodGet, "https://pkg.go.dev/"+module.Coordinates(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.Name, version.Version))

	query := req.URL.Query()
	query.Add("tab", "licenses")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		return "", ErrModuleNotFound
	default:
		return "", fmt.Errorf("unexpected response status: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	sel := doc.Find("div.Container section.License h2").First()
	if len(sel.Nodes) == 0 {
		return "", ErrLicenseNotFound
	}

	license := strings.TrimSpace(sel.Text())
	if license == "" {
		return "", ErrLicenseNotFound
	}

	return license, nil
}
