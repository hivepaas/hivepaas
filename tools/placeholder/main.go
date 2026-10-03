// Command placeholder is what a new app runs until its first deployment: it
// answers with a page saying so, and stops as soon as it is told to.
//
// An app is created before anything is deployed to it, and its service needs an
// image that stays up - a task that exits is restarted by swarm again and
// again. This one runs without a command of its own to override, so whatever
// image the app is given later runs its own. It stops on SIGTERM (a container
// stopped, or replaced by a deployment) and SIGINT, at once and with status 0.
//
// A new app has no port yet: its routing is set later, to the port its own
// image will listen on. So the page is served on the ports apps listen on most
// often (80, 3000, 8000, 8080); a domain routed to another port shows the
// proxy's error until the app is deployed.
//
// It is published as ghcr.io/hivepaas/placeholder, one static binary on an
// empty image (deployment/release/Dockerfile.placeholder), and release.json
// names it by digest as placeholderImage.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// defaultAddrs are the ports apps listen on most often.
var defaultAddrs = []string{":80", ":3000", ":8000", ":8080"}

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 2 * time.Second
)

// page is the whole answer: nothing of the request is written back.
const page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Not deployed yet</title>
<style>
:root { color-scheme: light dark; --bg: #f6f7f9; --fg: #1d2330; --muted: #5b6475; --card: #ffffff; --line: #e3e6ec; }
@media (prefers-color-scheme: dark) {
  :root { --bg: #0f1218; --fg: #e8ebf1; --muted: #9aa3b5; --card: #171b23; --line: #262c38; }
}
body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: var(--bg); color: var(--fg);
  font: 16px/1.6 system-ui, -apple-system, "Segoe UI", Roboto, sans-serif; padding: 0 16px; }
main { max-width: 32rem; background: var(--card); border: 1px solid var(--line); border-radius: 12px;
  padding: 2rem 2.25rem; }
h1 { font-size: 1.35rem; margin: 0 0 .5rem; text-wrap: balance; }
p { margin: .5rem 0; color: var(--muted); }
small { display: block; margin-top: 1.5rem; color: var(--muted); }
</style>
</head>
<body>
<main>
<h1>This app has not been deployed yet</h1>
<p>It is ready on HivePaaS and waits for its first deployment.</p>
<p>Set its image or its source in the app's Deployment Settings, and deploy it: this page will be replaced by the app.</p>
<small>HivePaaS</small>
</main>
</body>
</html>
`

func main() {
	run(defaultAddrs)
}

// run serves the page on addrs until SIGTERM or SIGINT.
func run(addrs []string) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	serve(listen(addrs), stop)
}

// listen listens on every address it can. One it cannot is skipped, and said
// so: the placeholder keeps running for the others, and keeps waiting with none.
func listen(addrs []string) []net.Listener {
	listeners := make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "placeholder: not listening on %s: %v\n", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

// serve answers on listeners until stop receives, then stops them.
func serve(listeners []net.Listener, stop <-chan os.Signal) {
	srv := &http.Server{Handler: http.HandlerFunc(welcome), ReadHeaderTimeout: readHeaderTimeout}
	for _, listener := range listeners {
		go func() { //safego:allow a standalone binary: a panic ends the container, which swarm restarts
			if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Fprintf(os.Stderr, "placeholder: %s: %v\n", listener.Addr(), err)
			}
		}()
	}
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// welcome answers every request with the page. It is not kept: once the app is
// deployed, the app answers.
func welcome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(page))
}
