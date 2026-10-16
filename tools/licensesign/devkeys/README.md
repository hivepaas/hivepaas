# Development license signing keys

These private keys are public, on purpose. They sign license keys for local
development and for the e2e environment (`hivepaas-dashboard/e2e/env/up.sh`),
which makes far more apps than the free plan allows:

```
go run ./tools/licensesign issue -key tools/licensesign/devkeys/dev-ed.key \
  -key tools/licensesign/devkeys/dev-ml.key -plan colony \
  -installation "$(docker info -f '{{.Swarm.Cluster.ID}}')"
```

Only a HivePaaS running with `HP_ENV=development` accepts a key they signed.
Anywhere else it is refused as a development key, so anybody being able to sign
one costs nothing.

Their public halves are `hivepaas_app/pkg/license/licensekeys/dev/`, and a test
checks the two match. They are committed despite the repository's `*.key`
ignore rule by the `.gitignore` here, which names them alone.

Never put a production license key here.
