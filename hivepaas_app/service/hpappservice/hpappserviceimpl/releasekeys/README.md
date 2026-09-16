# Release signing keys

The public keys `release.json` is accepted from. Every `*.pub.pem` file here is
compiled into the binary; the file name without `.pub.pem` is the key id that
`release.json.sig` refers to.

`release.json` must carry a valid **ed25519** and a valid **ML-DSA-65** signature,
each by a key in this directory (see `hivepaas_app/pkg/releasesig`). A binary built
without a key of either algorithm here cannot see or apply updates.

Create keys on the offline signing machine, with the tool built the way
`scripts/release-sign.sh` builds it:

```
releasesign keygen -alg ed25519   -key-id 2026-ed -dir /offline
releasesign keygen -alg ml-dsa-65 -key-id 2026-ml -dir /offline
```

and copy only the `.pub.pem` files here. The `.key` files never leave that machine.

## Rotating a key

A binary keeps trusting the keys it was built with, so:

1. Add the new key's `.pub.pem` here, next to the old one, and release.
2. Once that release is widely installed, sign with the new key (`make release-sign`).
   Older binaries skip the signature by a key they do not know and still require
   one by a key they do, so sign with both old and new keys for as long as older
   binaries matter.
3. Remove the old key's `.pub.pem` in a later release.

Never remove a key in the same release that stops signing with it.
