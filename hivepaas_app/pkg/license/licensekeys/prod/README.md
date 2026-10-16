# License signing keys

The public keys license keys are accepted from, in every installation. Every
`*.pub.pem` file here is compiled into the app; the file name without
`.pub.pem` is the key id that a license key's signatures refer to, such as
`lic-2026-ed` and `lic-2026-ml`. No id here may also be one in `../dev/`.

A license key must carry a valid **ed25519** and a valid **ML-DSA-65**
signature, each by a key in this directory (see `hivepaas_app/pkg/license`).
Until a key of each algorithm is here, an installation takes no license key
but a development one, and only in development.

These keys are not the release keys (`pkg/releasesig/releasekeys`): licenses
are signed by the portal, unattended, and a key it holds must never be able to
sign a release. They also sign under a context of their own,
`hivepaas-license-v1`.

Create a pair with `tools/licensesign`, where license keys are signed:

```
licensesign keygen -alg ed25519   -key-id lic-2026-ed -dir /secure
licensesign keygen -alg ml-dsa-65 -key-id lic-2026-ml -dir /secure
```

and copy only the `.pub.pem` files here. The `.key` files live only where the
portal signs license keys, and with whoever signs support keys by hand; never
in a repository.

## Rotating a key

A binary keeps trusting the keys it was built with, and an installation that
never updates keeps the keys it has. So:

1. Add the new pair's `.pub.pem` files here, next to the old ones, and release.
2. Only once that release is widely installed, let the portal sign with the new
   pair. A key signed by it is refused by every binary built before step 1.
3. Remove the old pair's files only when no key it signed is still in force: a
   yearly key holds for a year and its grace days after it is issued.

Never remove a key in the same release that stops signing with it.
