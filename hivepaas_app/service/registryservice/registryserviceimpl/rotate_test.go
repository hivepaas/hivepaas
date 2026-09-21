package registryserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func lineFor(t *testing.T, password string) string {
	t.Helper()

	line, err := htpasswdLine(password)
	if err != nil {
		t.Fatalf("htpasswdLine: %v", err)
	}
	return line
}

// bcrypt is salted, so the same password hashes differently every time. A save
// that rewrote the line would replace the swarm secret and restart the registry
// each time an operator pressed Save.
func TestHtpasswdForKeepsTheLineThatAlreadyVerifies(t *testing.T) {
	existing := htpasswdContent(lineFor(t, "the-password"))

	content, err := htpasswdFor(existing, "the-password", false)
	if err != nil {
		t.Fatalf("htpasswdFor: %v", err)
	}

	assert.Equal(t, existing, content)
}

func TestHtpasswdForWritesALineWhenThereIsNone(t *testing.T) {
	content, err := htpasswdFor("", "the-password", false)
	if err != nil {
		t.Fatalf("htpasswdFor: %v", err)
	}

	_, hash, found := strings.Cut(strings.TrimSpace(content), ":")
	assert.True(t, found)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("the-password")))
}

// During the grace period both credentials work: an app that has not been
// redeployed still presents the one it was deployed with.
func TestHtpasswdForKeepsThePreviousLineDuringGrace(t *testing.T) {
	existing := htpasswdContent(lineFor(t, "new-password"), lineFor(t, "old-password"))

	content, err := htpasswdFor(existing, "new-password", true)
	if err != nil {
		t.Fatalf("htpasswdFor: %v", err)
	}

	assert.Equal(t, 2, strings.Count(content, "\n"))
	_, oldHash, _ := strings.Cut(nonEmptyLines(content)[1], ":")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(oldHash), []byte("old-password")))
}

// Once the grace is over the old line goes, which is the point of having one.
func TestHtpasswdForDropsThePreviousLineAfterGrace(t *testing.T) {
	existing := htpasswdContent(lineFor(t, "new-password"), lineFor(t, "old-password"))

	content, err := htpasswdFor(existing, "new-password", false)
	if err != nil {
		t.Fatalf("htpasswdFor: %v", err)
	}

	assert.Equal(t, 1, strings.Count(content, "\n"))
	_, hash, _ := strings.Cut(strings.TrimSpace(content), ":")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("new-password")))
}

// A file whose only line is for a password nobody holds any more gets a new line
// for the one in use, and keeps the old one only while the grace lasts.
func TestHtpasswdForAfterTheCredentialWasReplacedElsewhere(t *testing.T) {
	existing := htpasswdContent(lineFor(t, "stale-password"))

	content, err := htpasswdFor(existing, "current-password", true)
	if err != nil {
		t.Fatalf("htpasswdFor: %v", err)
	}

	lines := nonEmptyLines(content)
	assert.Len(t, lines, 2)
	_, currentHash, _ := strings.Cut(lines[0], ":")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(currentHash), []byte("current-password")))
}
