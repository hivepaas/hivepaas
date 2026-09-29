// Command releaseverify checks release.signed.json for the installer.
//
// The installer cannot check the signatures itself: a server's openssl is
// older than ML-DSA and Ed25519ctx. It runs this, built into a small image the
// installer pins by digest (deployment/release/Dockerfile.verify), with the
// envelope on stdin:
//
//	docker run --rm -i --network none <image> < release.signed.json > release.json
//
// It trusts the app's own keys (pkg/releasesig/releasekeys) and requires what
// the app requires: a valid ed25519 and a valid ML-DSA-65 signature. It writes
// release.json to stdout only once both verify, and nothing otherwise, with
// exit status 1.
//
// With -fingerprint it prints the fingerprint of the keys it trusts, which is
// what install.sh records as VERIFY_KEYS beside the image's digest.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig/releasekeys"
)

// maxEnvelopeSize bounds what is read: release.signed.json is a few KB.
const maxEnvelopeSize = 4 << 20

func main() {
	pems, err := releasekeys.Embedded()
	if err != nil {
		fmt.Fprintln(os.Stderr, "releaseverify: reading the trusted keys:", err)
		os.Exit(1)
	}
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, pems))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, pems map[string][]byte) int {
	flags := flag.NewFlagSet("releaseverify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	fingerprint := flags.Bool("fingerprint", false, "print the fingerprint of the trusted keys and exit")
	if err := flags.Parse(args); err != nil {
		return 2 //nolint:mnd
	}
	if *fingerprint {
		_, _ = fmt.Fprintln(stdout, releasekeys.Fingerprint(pems))
		return 0
	}

	keys, err := releasesig.ParsePublicKeys(pems)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "releaseverify: the trusted keys:", err)
		return 1
	}
	envelope, err := io.ReadAll(io.LimitReader(stdin, maxEnvelopeSize+1))
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "releaseverify: reading the envelope:", err)
		return 1
	}
	if len(envelope) > maxEnvelopeSize {
		_, _ = fmt.Fprintln(stderr, "releaseverify: the envelope is larger than any release.signed.json")
		return 1
	}
	release, err := releasesig.Open(keys, envelope)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "releaseverify: release.signed.json does not verify:", err)
		return 1
	}
	if _, err = stdout.Write(release); err != nil {
		_, _ = fmt.Fprintln(stderr, "releaseverify: writing release.json:", err)
		return 1
	}
	return 0
}
