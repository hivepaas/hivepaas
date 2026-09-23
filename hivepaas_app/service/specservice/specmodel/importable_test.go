package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestCheckImportableAcceptsWhatATemplateCannotSay(t *testing.T) {
	doc := decodeDoc(t, `
deployment:
  container: {user: "999", readOnly: true}
  service: {modeSpec: {mode: global}}
  resources: {memory: {swap: 1gb}, reservations: {genericResources: [{kind: gpu, value: '1'}]}}
  networks: {dnsConfig: {nameservers: [1.1.1.1]}, attachments: [{name: shop_prod}]}
  storage:
    mounts: {/data: {type: cluster, source: vol-1, clusterOptions: {subpath: data}}}
    dockerMounts: {/cache: {type: tmpfs}}
`)
	assert.NoError(t, CheckImportable(doc))
}

func TestCheckImportableRefuses(t *testing.T) {
	cases := map[string]string{
		"a volume import has not resolved": "deployment:\n  storage:\n    mounts:\n" +
			"      /data: {type: volume, external: {type: cluster-volume, name: v}}\n",
		"a managed mount with no volume": "deployment:\n  storage:\n    mounts:\n      /data: {type: volume}\n",
		"a managed bind":                 "deployment:\n  storage:\n    mounts:\n      /data: {type: bind, source: v}\n",
		"a relative target":              "deployment:\n  storage:\n    dockerMounts:\n      data: {type: tmpfs}\n",
		"a target in both maps": "deployment:\n  storage:\n    mounts:\n      /d: {type: volume, source: v}\n" +
			"    dockerMounts:\n      /d: {type: tmpfs}\n",
		"a subpath leaving the directory": "deployment:\n  storage:\n    mounts:\n" +
			"      /d: {type: volume, source: v, volumeOptions: {subpath: ../x}}\n",
		"an unknown mode":  "deployment:\n  service: {modeSpec: {mode: sometimes}}\n",
		"a port too large": "deployment:\n  networks: {endpointSpec: {ports: [{target: 70000}]}}\n",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, CheckImportable(decodeDoc(t, doc)), hperrors.ErrSpecBlockInvalid)
		})
	}
}

// Every deployment block is built for an export with a deployment - a block it
// left out is one the app has none of - the source only when it is there, and
// the settings as one block.
func TestImportBlocks(t *testing.T) {
	assert.Equal(t, []Block{
		BlockDeploymentStorage, BlockContainer, BlockDeploymentResources, BlockDeploymentNetworks,
		BlockDeploymentService, BlockSettings,
	}, ImportBlocks(decodeDoc(t, "deployment:\n  container: {}\nsettings:\n  kind: {category: webapp}\n")))
	assert.Equal(t, []Block{BlockSettings},
		ImportBlocks(decodeDoc(t, "settings:\n  routing: {port: 80}\n")), "an app never deployed")
	assert.Empty(t, ImportBlocks(&AppDoc{}))
}
