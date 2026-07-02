package translation

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// load loads data files of the given lang
func load(rootPath string, lang Lang) (*i18n.Localizer, error) {
	locale := localeMap[lang]
	bundle := i18n.NewBundle(locale)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load all data files in the dir
	dataPath := path.Join(rootPath, lang.String())
	if err := walkOverMsgFiles(dataPath, func(dir string, e fs.DirEntry) error {
		msgFile := path.Join(dir, e.Name())
		filebytes, err := messages.ReadFile(msgFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", msgFile, err)
		}
		if _, err = bundle.ParseMessageFileBytes(filebytes, e.Name()); err != nil {
			return fmt.Errorf("failed to parse file %s: %w", msgFile, err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return i18n.NewLocalizer(bundle, lang.String()), nil
}

// walkOverMsgFiles travels through the given directory (recursively)
func walkOverMsgFiles(dir string, parseFn func(dir string, e fs.DirEntry) error) error {
	entries, err := messages.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			// If directory, load data from it recursively
			if err := walkOverMsgFiles(path.Join(dir, e.Name()), parseFn); err != nil {
				return err
			}
			continue
		}
		if filepath.Ext(e.Name()) == ".toml" {
			// Accept file with expected extension only
			if err := parseFn(dir, e); err != nil {
				return err
			}
		}
	}
	return nil
}
