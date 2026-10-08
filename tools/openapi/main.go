// Command openapi keeps docs/openapi/swagger.json true to the code it describes.
//
// Swag reads the API from handler comments and Go types. What it reads from the
// types is right; what the comments say - the routes, the query parameters - is
// written by hand and can fall behind, and some of what the types mean is lost
// on the way. This tool covers both:
//
//	fix   corrects the generated spec from the Go types: a field with omitempty
//	      is optional, one that can be null is nullable, one of type any takes
//	      any value. tools/swag/swag.sh runs it after the conversion to OpenAPI 3.
//	level keeps the API level a HivePaaS CLI must be built for to write, raised
//	      when the request of an existing write operation changes. swag.sh runs
//	      it after fix.
//	lint  checks every handler's comment against what the handler reads: each
//	      query parameter its request type decodes is documented, each one
//	      documented is read, and the path parameters match the route.
//
// That every route is in the spec, and nothing else, is a test in
// hivepaas_app/interface/api/server, where the routes are.
//
// Usage:
//
//	go run ./tools/openapi fix   [file]   (default docs/openapi/swagger.json)
//	go run ./tools/openapi level [file]
//	go run ./tools/openapi lint
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const defaultSpec = "docs/openapi/swagger.json"

var errProblems = errors.New("the API handlers' comments do not match the code")

func main() {
	if len(os.Args) < 2 { //nolint:mnd
		usage()
	}
	var err error
	switch os.Args[1] {
	case "fix":
		err = runFix(os.Args[2:], os.Stdout)
	case "level":
		err = runLevel(os.Args[2:], os.Stdout)
	case "lint":
		err = runLint(os.Args[2:], os.Stdout)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "openapi:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: openapi fix [file] | openapi level [file] | openapi lint")
	os.Exit(2) //nolint:mnd
}

func loadModule() (*goSource, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	root, module, err := moduleRoot(wd)
	if err != nil {
		return nil, "", fmt.Errorf("no go.mod above %s", wd)
	}
	src, err := loadGoSource(root, module, "hivepaas_app")
	return src, root, err
}

func runFix(args []string, out io.Writer) error {
	src, root, err := loadModule()
	if err != nil {
		return err
	}
	file := filepath.Join(root, defaultSpec)
	if len(args) > 0 {
		file = args[0]
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	value, err := decodeDocument(data)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	doc, ok := value.(*object)
	if !ok {
		return fmt.Errorf("%s: not a JSON object", file)
	}
	stats, err := fixSpec(doc, src)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	if err = os.WriteFile(file, encodeDocument(doc), 0o644); err != nil { //nolint:gosec
		return err
	}
	fmt.Fprintf(out, "%s: %s\n", file, stats)
	return nil
}
