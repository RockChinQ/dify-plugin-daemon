package bundle_packager

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type LocalBundlePackager struct {
	GenericBundlePackager

	path string
}

func NewLocalBundlePackager(path string) (BundlePackager, error) {
	// try read manifest file
	manifestFile, err := os.Open(filepath.Join(path, "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	defer manifestFile.Close()

	manifestBytes, err := io.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	bundle, err := parser.UnmarshalYamlBytes[bundle_entities.Bundle](manifestBytes)
	if err != nil {
		return nil, err
	}

	packager := &LocalBundlePackager{
		GenericBundlePackager: *NewGenericBundlePackager(&bundle),
		path:                  path,
	}

	// walk through the path and load the assets
	err = filepath.Walk(filepath.Join(path, "_assets"), func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		assetBytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		prefix := filepath.Join(path, "_assets")
		assetName := strings.TrimPrefix(filePath, prefix)
		packager.assets[assetName] = bytes.NewBuffer(assetBytes)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return packager, nil
}

func (p *LocalBundlePackager) Save() error {
	// save the assets
	for name, asset := range p.assets {
		err := os.WriteFile(filepath.Join(p.path, "_assets", name), asset.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	// save the manifest file
	manifestBytes := parser.MarshalYamlBytes(p.bundle)
	err := os.WriteFile(filepath.Join(p.path, "manifest.yaml"), manifestBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}