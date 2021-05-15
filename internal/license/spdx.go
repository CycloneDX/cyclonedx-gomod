package license

import (
	_ "embed"
	"encoding/json"
	"log"
)

var (
	//go:embed spdx-licenses.json
	licensesJSON []byte

	licenses []SPDXLicense
)

type SPDXLicense struct {
	ID        string `json:"licenseId"`
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

func init() {
	licensesObj := struct {
		Licenses []SPDXLicense `json:"licenses"`
	}{}
	if err := json.Unmarshal(licensesJSON, &licensesObj); err != nil {
		log.Fatalf("failed to unmarshal SPDX license list: %v", err)
	}
	licenses = licensesObj.Licenses
}

func getLicenseByID(licenseID string) *SPDXLicense {
	for i := range licenses {
		if licenses[i].ID == licenseID {
			return &licenses[i]
		}
	}
	return nil
}
