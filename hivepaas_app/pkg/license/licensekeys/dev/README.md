# Development license keys

The public keys of the development key pair. Their private halves are
committed, in `tools/licensesign/devkeys/`, so anybody can sign a license key
with them: for local development and for the e2e environment, which makes far
more apps than the free plan allows.

A key they signed is accepted only by a HivePaaS running with
`HP_ENV=development`. Anywhere else it is refused as a development key.

`tools/licensesign`'s tests check that the committed private keys derive
exactly these files. To replace the pair, generate it into
`tools/licensesign/devkeys/` with `licensesign keygen` (`-key-id dev-ed` and
`-key-id dev-ml`) and move the `.pub.pem` files here.
