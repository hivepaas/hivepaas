package base

import "github.com/tiendc/gofn"

// SecretType names a kind of secret the API can hand out in the clear.
//
// It is not SettingType. A setting is a record somebody stored; these are the
// distinct kinds of secret a reveal can be about, and most of them - every stored
// setting - share the empty value because the operator flag applies to them as
// one. A type gets named here only when it needs to be exempt from that flag
// separately, which is a decision about the secret's nature rather than about
// where it happens to live.
type SecretType string

const (
	// SecretTypeSwarmJoinToken is the token that lets a machine into the cluster.
	//
	// Named because it is not a stored secret at all - it comes from the Swarm
	// itself - so the flag that decides whether the database's secrets may leave
	// the server is not obviously the right gate for it, and an operator may
	// reasonably want node onboarding to work while the rest stays shut.
	SecretTypeSwarmJoinToken SecretType = "swarm-join-token"
)

// AllSecretTypes are the types that may be listed as exempt from
// ReturnSecretsViaAPI. The empty type is absent on purpose: it stands for "a
// stored secret", and exempting all of those is what the flag already does.
//
// Before adding to this list, ask which of two things the secret is.
//
// A capability lets the holder do something, and the system can take it back:
// the Swarm join token is one, and `docker swarm join-token --rotate` closes a
// leak completely. A key decrypts data, and nothing takes it back: the backup
// repository password is one, and changing it does not re-encrypt the snapshots
// already written, so a single leak exposes the whole history for good.
//
// Only capabilities belong here. The argument that a leaked key is "just more
// risk" misses that the risk cannot be retired - there is no action an operator
// can take afterwards that undoes it.
//
// Two supporting tests for a candidate, both of which the join token passes and
// a stored key does not. Is the secret needed for an ordinary operation with no
// other path, rather than for a rare one-off? And does retrieving it while the
// system is healthy actually serve the case it is wanted for - a key wanted for
// disaster recovery is wanted exactly when this API is unreachable, so exempting
// it widens the attacker's window without widening the operator's.
var AllSecretTypes = []SecretType{
	SecretTypeSwarmJoinToken,
}

// IsValidSecretType reports whether value names a type that can be exempted.
//
// The check exists so a typo in the config is refused rather than silently
// exempting nothing - an operator who believes they have opened a door and has
// not is worse off than one who was told the name was wrong.
func IsValidSecretType(value string) bool {
	return gofn.Contain(AllSecretTypes, SecretType(value))
}
