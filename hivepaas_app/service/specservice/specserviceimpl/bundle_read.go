package specserviceimpl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"path"
	"strings"

	"filippo.io/age"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	// ageHeader starts every age file, which is how an encrypted bundle is told
	// from a plain one without asking.
	ageHeader = "age-encryption.org/"

	// A bundle is configuration: an app is some hundreds of lines. The caps are
	// far above any real installation and exist so that a crafted archive cannot
	// make the process inflate an unbounded amount into memory.
	maxBundleFiles     = 10000
	maxBundleFileBytes = 16 << 20
	maxBundleBytes     = 128 << 20
)

// readBundle reads a bundle as it was uploaded: decrypts it when it is an age
// file, unpacks the archive, and parses every document in it.
func readBundle(content []byte, passphrase string) (*specmodel.ImportBundle, error) {
	digest := sha256.Sum256(content)

	if bytes.HasPrefix(content, []byte(ageHeader)) {
		if passphrase == "" {
			return nil, hperrors.Wrap(hperrors.ErrSpecPassphraseRequired)
		}
		decrypted, err := decryptBundle(content, passphrase)
		if err != nil {
			return nil, err
		}
		content = decrypted
	}

	files, err := untarGz(content)
	if err != nil {
		return nil, err
	}
	bundle, err := parseBundleFiles(files)
	if err != nil {
		return nil, err
	}
	bundle.Digest = hex.EncodeToString(digest[:])
	return bundle, nil
}

func decryptBundle(content []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSpecPassphraseInvalid)
	}
	reader, err := age.Decrypt(bytes.NewReader(content), identity)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSpecPassphraseInvalid)
	}
	decrypted, err := io.ReadAll(io.LimitReader(reader, maxBundleBytes+1))
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSpecBundleInvalid).WithExtraDetail("%s", err.Error())
	}
	return decrypted, nil
}

// untarGz reads every regular file of a gzipped tar into memory, by its path
// inside the archive. A path that is absolute or climbs out of the archive is
// refused rather than cleaned: nothing export writes looks like that.
func untarGz(content []byte) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		return nil, invalidBundle("not a gzip archive")
	}
	defer gz.Close()

	files := map[string][]byte{}
	total := 0
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, invalidBundle("not a tar archive")
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		name := strings.TrimPrefix(header.Name, "./")
		// macOS's tar stores extended attributes as AppleDouble files beside the
		// real ones, named ._<file>. They are metadata, not documents.
		if strings.HasPrefix(path.Base(name), "._") {
			continue
		}
		if path.IsAbs(name) || name != path.Clean(name) || name == ".." || strings.HasPrefix(name, "../") {
			return nil, invalidBundle("the entry %q leaves the archive", header.Name)
		}
		if len(files) >= maxBundleFiles {
			return nil, invalidBundle("more than %d files", maxBundleFiles)
		}
		data, err := io.ReadAll(io.LimitReader(reader, maxBundleFileBytes+1))
		if err != nil {
			return nil, invalidBundle("%s cannot be read", name)
		}
		if len(data) > maxBundleFileBytes {
			return nil, invalidBundle("%s is larger than %d bytes", name, maxBundleFileBytes)
		}
		if total += len(data); total > maxBundleBytes {
			return nil, invalidBundle("larger than %d bytes unpacked", maxBundleBytes)
		}
		files[name] = data
	}
	return files, nil
}

// parseBundleFiles parses the documents of a bundle. It reads what an upload
// holds and what export builds for this installation alike, which is what makes
// the two comparable.
func parseBundleFiles(files map[string][]byte) (*specmodel.ImportBundle, error) {
	raw, ok := files[manifestFilename]
	if !ok {
		return nil, invalidBundle("there is no %s", manifestFilename)
	}
	bundle := &specmodel.ImportBundle{
		Manifest: &specmodel.Manifest{},
		Projects: map[string]*specmodel.ProjectDoc{},
		Envs:     map[string]map[string]*specmodel.EnvDoc{},
	}
	if err := yaml.Unmarshal(raw, bundle.Manifest); err != nil {
		return nil, invalidBundle("%s: %s", manifestFilename, err.Error())
	}
	if bundle.Manifest.APIVersion != specmodel.APIVersion {
		return nil, hperrors.Wrap(hperrors.ErrSpecAPIVersionUnsupported).
			WithExtraDetail("%s", bundle.Manifest.APIVersion)
	}

	for name, content := range files {
		parts := strings.Split(name, "/")
		switch {
		case name == globalFilename:
			bundle.Global = &specmodel.GlobalDoc{}
			if err := parseDoc(name, content, bundle.Global); err != nil {
				return nil, err
			}
		case len(parts) == 3 && parts[0] == projectsSegment && parts[2] == projectFilename:
			doc := &specmodel.ProjectDoc{}
			if err := parseDoc(name, content, doc); err != nil {
				return nil, err
			}
			bundle.Projects[parts[1]] = doc
		case len(parts) == 4 && parts[0] == projectsSegment && parts[2] == "envs" && strings.HasSuffix(parts[3], ".yaml"):
			doc := &specmodel.EnvDoc{}
			if err := parseDoc(name, content, doc); err != nil {
				return nil, err
			}
			if bundle.Envs[parts[1]] == nil {
				bundle.Envs[parts[1]] = map[string]*specmodel.EnvDoc{}
			}
			bundle.Envs[parts[1]][strings.TrimSuffix(parts[3], ".yaml")] = doc
		}
	}
	return bundle, nil
}

func parseDoc(name string, content []byte, out any) error {
	if err := yaml.Unmarshal(content, out); err != nil {
		return invalidBundle("%s: %s", name, err.Error())
	}
	return nil
}

func invalidBundle(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrSpecBundleInvalid).WithExtraDetail(format, args...)
}
