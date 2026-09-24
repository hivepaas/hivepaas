package volumeservice

import (
	"fmt"

	"github.com/kballard/go-shellquote"
)

// MakeDirWritableCmd is the shell that creates a directory an app's data is kept
// in and, while it is still empty, lets a container running as any user write
// to it.
//
// Only an empty directory is opened up. The directory is created by a helper
// running as root, so without this an app that does not run as root could not
// write to its own storage. Once the app has put something there, it owns what
// it made and has set the modes it wants - a database's 0700 data directory, a
// 0600 key, an .ssh directory sshd refuses to trust if anyone else can write to
// it - and opening those up on every deployment is how they used to be undone.
// It also kept every deployment waiting on a walk of the whole tree, which on a
// directory of cloned repositories is most of the deployment.
//
// Nothing below the directory is touched, so an app that has to be given data
// another user wrote - a different image on the same volume, files copied in by
// hand - is given it by resetting the mount's permissions, on request.
func MakeDirWritableCmd(dir string) string {
	q := shellquote.Join(dir)
	return fmt.Sprintf(`mkdir -p %s && { [ -n "$(ls -A %s)" ] || chmod 777 %s; }`, q, q, q)
}
