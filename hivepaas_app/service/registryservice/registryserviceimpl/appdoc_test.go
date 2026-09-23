package registryserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func docInput() appDocInput {
	return appDocInput{
		Name:        "Registry",
		Key:         "registry",
		Domain:      "registry.example.com",
		MemoryLimit: "512mb",
		OomScoreAdj: -300,
		VolumeID:    "01M32ATCXAVPK6JYV0HNCFCJYM",
		ZotConfig:   "{\n  \"distSpecVersion\": \"1.1.1\"\n}",
		Htpasswd:    "hivepaas:$2y$05$abc\n",
	}
}

func TestAppDocIsBuildable(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	// CheckBuildable is what BuildApp runs first; a document that fails it would
	// fail at provisioning time instead, with the app half created.
	assert.NoError(t, specmodel.CheckBuildable(doc))

	// The document carries no app or name: CheckBuildable refuses any top-level
	// field but deployment and settings, and the app's identity comes from the
	// provisioning request.
	assert.Empty(t, doc.App)

	// Source is an untyped block: it is the assembled app-deployment setting, and
	// build_deployment.go decodes it. Reading it the same way is what checks the
	// image actually landed in it.
	source, _ := doc.Deployment.Source["imageSource"].(map[string]any)
	assert.Equal(t, registryImage, source["image"])
	assert.Equal(t, "image", doc.Deployment.Source["activeMethod"])
}

// The registry outranks user apps when memory runs out, which is safe only
// because it cannot grow past its memory limit.
func TestAppDocProtectsTheRegistryFromTheOOMKiller(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	res := doc.Deployment.Resources
	if res == nil || res.Limits == nil || res.Capabilities == nil {
		t.Fatalf("want limits and capabilities, got %+v", res)
	}
	assert.Equal(t, "512mb", res.Limits.Memory.String())
	assert.Equal(t, int64(-300), res.Capabilities.OomScoreAdj)
}

func TestAppDocMountsTheVolume(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	mnt, ok := doc.Deployment.Storage.Mounts[registryRootDir]
	assert.True(t, ok, "the images have to be on the volume")
	assert.Equal(t, "01M32ATCXAVPK6JYV0HNCFCJYM", mnt.Source)
}

// With S3 there is nothing local worth keeping: zot rebuilds its cache from the
// bucket, and a mount would pin the app to a node for no reason.
func TestAppDocWithoutAVolumeHasNoMount(t *testing.T) {
	in := docInput()
	in.VolumeID = ""

	doc, err := renderAppDoc(in)
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	assert.Nil(t, doc.Deployment.Storage)
}

func TestAppDocCarriesTheConfigAndTheAccount(t *testing.T) {
	doc, err := renderAppDoc(docInput())
	if err != nil {
		t.Fatalf("renderAppDoc: %v", err)
	}

	configFiles, _ := doc.Settings["configFiles"].(map[string]any)
	assert.Contains(t, configFiles, "config.json")

	secrets, _ := doc.Settings["secrets"].(map[string]any)
	assert.Contains(t, secrets, "ZOT_HTPASSWD")

	routing, _ := doc.Settings["routing"].(map[string]any)
	assert.Equal(t, 5000, routing["port"])
}
