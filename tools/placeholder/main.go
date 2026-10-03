// Command placeholder is what a new app runs until its first deployment: it
// waits, and stops as soon as it is told to.
//
// An app is created before anything is deployed to it, and its service needs an
// image that stays up - a task that exits is restarted by swarm again and
// again. This one waits without a command of its own to override, so whatever
// image the app is given later runs its own. It stops on SIGTERM (a container
// stopped, or replaced by a deployment) and SIGINT, at once and with status 0.
//
// It is published as ghcr.io/hivepaas/placeholder, one static binary on an
// empty image (deployment/release/Dockerfile.placeholder), and release.json
// names it by digest as placeholderImage.
package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
}
