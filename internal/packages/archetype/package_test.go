// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package archetype

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/elastic-package/internal/packages"
	"github.com/elastic/elastic-package/internal/validation"
)

func TestPackage(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pd := createPackageDescriptorForTest("integration", "^7.13.0")
		createAndCheckPackage(t, pd, true)
	})
	t.Run("missing-version", func(t *testing.T) {
		pd := createPackageDescriptorForTest("integration", "^7.13.0")
		pd.Manifest.Version = ""
		createAndCheckPackage(t, pd, false)
	})
	t.Run("input-package", func(t *testing.T) {
		pd := createPackageDescriptorForTest("input", "^8.9.0")
		createAndCheckPackage(t, pd, true)
	})
}

func createAndCheckPackage(t *testing.T, pd PackageDescriptor, valid bool) {
	tempDir := t.TempDir()
	err := createPackageInDir(pd, tempDir)
	require.NoError(t, err)

	checkPackage(t, filepath.Join(tempDir, pd.Manifest.Name), valid)
}

func createPackageDescriptorForTest(packageType, kibanaVersion string) PackageDescriptor {
	var elasticsearch *packages.Elasticsearch
	inputDataStreamType := ""
	if packageType == "input" {
		inputDataStreamType = "logs"
		elasticsearch = &packages.Elasticsearch{
			IndexTemplate: &packages.ManifestIndexTemplate{
				Mappings: &packages.ManifestMappings{
					Subobjects: false,
				},
			},
		}
	}
	specVersion, err := GetLatestStableSpecVersion()
	if err != nil {
		panic(err)
	}
	return PackageDescriptor{
		Manifest: packages.PackageManifest{
			SpecVersion: specVersion.String(),
			Name:        "go_unit_test_package",
			Title:       "Go Unit Test Package",
			Type:        packageType,
			Version:     "1.2.3",
			Conditions: packages.Conditions{
				Kibana: packages.KibanaConditions{
					Version: kibanaVersion,
				},
				Elastic: packages.ElasticConditions{
					Subscription: "basic",
				},
			},
			Owner: packages.Owner{
				Github: "mtojek",
				Type:   "elastic",
			},
			Description:   "This package has been generated by a Go unit test.",
			Categories:    []string{"aws", "custom"},
			Elasticsearch: elasticsearch,
		},
		InputDataStreamType: inputDataStreamType,
	}
}

func checkPackage(t *testing.T, packageRoot string, valid bool) {
	err := validation.ValidateFromPath(packageRoot)
	if !valid {
		assert.Error(t, err)
		return
	}
	require.NoError(t, err)

	manifest, err := packages.ReadPackageManifestFromPackageRoot(packageRoot)
	require.NoError(t, err)

	// Running in subtests because manifest subobjects can be pointers that can panic when dereferenced by assertions.
	if manifest.Type == "input" {
		t.Run("input", func(t *testing.T) {
			t.Run("subobjects expected to false", func(t *testing.T) {
				assert.False(t, manifest.Elasticsearch.IndexTemplate.Mappings.Subobjects)
			})
		})
	}

	if manifest.Type == "integration" {
		t.Run("integration", func(t *testing.T) {
			ds, err := filepath.Glob(filepath.Join(packageRoot, "data_stream", "*"))
			require.NoError(t, err)
			for _, d := range ds {
				manifest, err := packages.ReadDataStreamManifest(filepath.Join(d, "manifest.yml"))
				require.NoError(t, err)
				t.Run("subobjects expected to false", func(t *testing.T) {
					assert.False(t, manifest.Elasticsearch.IndexTemplate.Mappings.Subobjects)
				})
			}
		})
	}
}
