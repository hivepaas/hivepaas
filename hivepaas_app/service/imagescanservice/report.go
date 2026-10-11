package imagescanservice

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// trivyReport is the part of `trivy image --format json` a scan keeps.
type trivyReport struct {
	Trivy struct {
		Version string `json:"Version"`
		Server  *struct {
			VulnerabilityDB *struct {
				UpdatedAt time.Time `json:"UpdatedAt"`
			} `json:"VulnerabilityDB"`
		} `json:"Server"`
	} `json:"Trivy"`
	ArtifactName string `json:"ArtifactName"`
	Metadata     struct {
		ImageID string `json:"ImageID"`
		OS      *struct {
			Family string `json:"Family"`
			Name   string `json:"Name"`
			EOSL   bool   `json:"EOSL"`
		} `json:"OS"`
		RepoTags    []string `json:"RepoTags"`
		RepoDigests []string `json:"RepoDigests"`
		ImageConfig struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"ImageConfig"`
	} `json:"Metadata"`
	Results []struct {
		Target          string `json:"Target"`
		Type            string `json:"Type"`
		Vulnerabilities []struct {
			VulnerabilityID  string     `json:"VulnerabilityID"`
			PkgName          string     `json:"PkgName"`
			PkgPath          string     `json:"PkgPath"`
			InstalledVersion string     `json:"InstalledVersion"`
			FixedVersion     string     `json:"FixedVersion"`
			Severity         string     `json:"Severity"`
			Title            string     `json:"Title"`
			PrimaryURL       string     `json:"PrimaryURL"`
			PublishedDate    *time.Time `json:"PublishedDate"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

// Report is a scan as it is kept.
type Report struct {
	Scan     *entity.ImageScan
	Findings []*entity.ImageScanFinding
}

// ParseReport reads a Trivy report into a scan of an image and its findings.
// imageRef is the reference the image was asked for by, shown when the report
// names no tag of its own.
func ParseReport(data []byte, imageRef string, scannedAt time.Time) (*Report, error) {
	var r trivyReport
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if r.Metadata.ImageID == "" {
		return nil, hperrors.Wrap(hperrors.ErrInfraInternal).WithParam("Error", "the report names no image")
	}

	scan := &entity.ImageScan{
		ImageID:      r.Metadata.ImageID,
		ImageRef:     imageRef,
		Status:       base.ImageScanStatusDone,
		TrivyVersion: r.Trivy.Version,
		ScannedAt:    scannedAt,
		UpdatedAt:    scannedAt,
	}
	if len(r.Metadata.RepoTags) > 0 && (imageRef == "" || strings.HasPrefix(imageRef, "sha256:")) {
		scan.ImageRef = r.Metadata.RepoTags[0]
	}
	if cfg := r.Metadata.ImageConfig; cfg.OS != "" {
		scan.Platform = cfg.OS + "/" + cfg.Architecture
	}
	if os := r.Metadata.OS; os != nil {
		scan.OSFamily, scan.OSName, scan.OSEOSL = os.Family, os.Name, os.EOSL
	}
	if server := r.Trivy.Server; server != nil && server.VulnerabilityDB != nil {
		scan.DBUpdatedAt = server.VulnerabilityDB.UpdatedAt
	}

	var findings []*entity.ImageScanFinding
	seen := map[string]bool{}
	for _, result := range r.Results {
		for _, v := range result.Vulnerabilities {
			finding := &entity.ImageScanFinding{
				ImageID:          scan.ImageID,
				VulnID:           v.VulnerabilityID,
				PkgName:          v.PkgName,
				PkgType:          result.Type,
				PkgPath:          v.PkgPath,
				InstalledVersion: v.InstalledVersion,
				FixedVersion:     v.FixedVersion,
				Severity:         severityOf(v.Severity),
				Title:            v.Title,
				URL:              v.PrimaryURL,
			}
			if v.PublishedDate != nil {
				finding.PublishedAt = *v.PublishedDate
			}
			// Trivy reports a package once per place it is found; one row each.
			if key := finding.Key(); !seen[key] {
				seen[key] = true
				findings = append(findings, finding)
			}
		}
	}
	return &Report{Scan: scan, Findings: findings}, nil
}

// severityOf is Trivy's grade as one of ours; UNKNOWN for one it is not.
func severityOf(s string) base.VulnSeverity {
	severity := base.VulnSeverity(strings.ToUpper(s))
	for _, known := range base.AllVulnSeverities {
		if severity == known {
			return severity
		}
	}
	return base.VulnSeverityUnknown
}
