# Letting something outside HivePaaS pull from the registry

**Status:** draft, awaiting review

## 1. The problem

The system registry has exactly one account: `hivepaas`, created with the registry, holding the
password every build pushes with. Anyone who needs an image from outside - a CI job that
deploys elsewhere, a developer pulling a build to run locally, a second cluster - has to be
given that account, which can push to every repository in the installation and cannot be taken
away from one consumer without taking it from all of them.

Two things measured on zot v2.1.21 say how to fix it, and how not to:

- **Its API keys are not the answer.** zot can create, list and delete them (`POST`/`GET`/
  `DELETE /zot/auth/apikey`, no update - editing means deleting and creating). But a key's
  `scopes` field is **not enforced**: a key scoped to `ci/x` started an upload in `ci/y` and
  pushed a whole image there. A key can mint further keys, and revoking the first leaves the
  second working. A key outlives its user: removing the user from htpasswd left the key
  answering 200. And an admin cannot list or revoke another user's keys.
- **`accessControl` is.** A user `ci` given `read, create, update` on `ci/**` pushed to
  `ci/build`, was refused `hivepaas/sneaky` with 403, could not read `hivepaas/app`, and saw
  only `ci/build` in `/v2/_catalog`.

So the unit of access is a **user**, not a token, and the enforcement is zot's `accessControl`.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| What is a credential? | An htpasswd user HivePaaS owns, with a generated password | Its password is the token. Deleting the line revokes it for good, which a zot API key does not |
| zot's API keys | Stay off (`http.auth.apikey` absent) | With them on, anyone holding a credential can mint keys that survive its deletion - measured |
| What may a credential reach? | Whole projects, or single apps, chosen from lists | The repository is `<project>-<app>`, so both are a pattern: `hivepaas/<project>-*` and `hivepaas/<project>-<app>`. No free text in this version |
| What may it do? | Pull, or pull and push | `read` against `read, create, update` in zot's terms. Delete stays with the system account, because cleanup is HivePaaS's job |
| Per environment? | No | `accessControl` matches repositories, never tags, and the environment is in the tag. §7 of the naming design records the same trade |
| Where is the password? | Encrypted in the setting, revealed the way every other secret is | HivePaaS re-renders the account file on every change and needs the hash; the operator needs the password more than once, and every other credential in the product can be revealed |
| What are they stored as? | A new collection setting type, `registry-credential`, at global scope | They are not registry auths: a registry auth is what HivePaaS pushes **with**, this is an account the registry **checks**. Keeping them apart keeps the app deployment screen's list clean |
| When do changes take effect? | On save, with a restart of a few seconds | The account file and the configuration are swarm objects, which are immutable |

## 3. The setting

`base.SettingTypeRegistryCredential = "registry-credential"`, a collection at global scope -
several per installation, each with a name - registered in `specmodel.collectionBlockNames` as
`registryCredentials` and given a spec policy that exports it whole.

```go
type RegistryCredential struct {
	// Username is what the consumer signs in with. It is also what the account
	// file and the access rules key on, so it is fixed after creation.
	Username string         `json:"username"`
	Password EncryptedField `json:"password"`

	// Access is what this credential may reach. Empty means nothing, which is a
	// credential that authenticates and is refused everything - the safe state
	// for one whose apps were all deleted.
	Access []RegistryCredentialAccess `json:"access,omitempty"`

	// Bcrypt is the hash the account file carries. It is stored because the file
	// is rendered on every change and rehashing the same password would rewrite
	// it every time, restarting the registry for nothing.
	Bcrypt string `json:"bcrypt"`
}

type RegistryCredentialAccess struct {
	// Project is the project whose images this reaches. Required.
	Project ObjectID `json:"project"`
	// App narrows it to one app of that project. Empty means every app in it.
	App ObjectID `json:"app,omitzero"`
	// Push says this may also push, not only pull.
	Push bool `json:"push,omitempty"`
}
```

The registry setting gains nothing: a credential names the project and the app, and the
patterns are derived when the configuration is rendered, so an app renamed or deleted changes
what the credential reaches without anybody editing it.

## 4. What zot is given

The configuration rendered for the registry app grows an `accessControl` block:

```jsonc
"accessControl": {
  "repositories": {
    "hivepaas/shop-*":   { "policies": [{"users": ["ci-github"], "actions": ["read"]}],
                           "defaultPolicy": [] },
    "hivepaas/blog-api": { "policies": [{"users": ["deployer"],
                                         "actions": ["read", "create", "update"]}],
                           "defaultPolicy": [] },
    "**":                { "policies": [], "defaultPolicy": [] }
  },
  "adminPolicy": { "users": ["hivepaas"], "actions": ["read", "create", "update", "delete"] }
}
```

Three rules hold it together:

- **`hivepaas` is in `adminPolicy` and nothing else may be.** Every build pushes as that
  account; a configuration that forgets it is a registry HivePaaS locks itself out of. The
  renderer writes it unconditionally and a test guards it.
- **`**` allows nobody.** A repository no rule names is reachable by the system account alone,
  so a credential can never be widened by accident.
- **Patterns come from the access list**, through the same `ImageRepoName` the build uses, so a
  credential and a push can never disagree about what a repository is called.

The account file is the system account's line plus one line per credential, in a stable order
so that saving without changes rewrites nothing.

## 5. The flow

Saving a credential, like everything else about the registry, goes through `registryservice`:

1. **Validate.** A username that is `[a-z0-9][a-z0-9_-]{1,39}` and is neither `hivepaas` nor
   taken; at least one access entry; each naming a project the caller may read, and an app of
   that project when it names one.
2. **On creation**, generate the password, hash it, store both. On **rotation**, generate
   again; there is no grace period here, because nothing carries these credentials but the
   consumer the operator hands them to - a swarm service never holds one.
3. **Render** the account file and the configuration from every active credential, and write
   them through the same config-file and secret update path the registry already uses. The
   registry restarts; a pull landing in that window is retried by whatever was pulling.
4. **Deleting** a credential removes its line and its rules on the next render. A consumer
   holding it is refused at once - that is the whole point of not using zot's API keys.

## 6. The dashboard

A table under the registry's settings page, below the status section, because these belong to
the registry rather than to a project:

```
Credentials                                                     [ Add credential ]

ci-github     pull          shop (all apps)            created 2 days ago   [Reveal] [Edit] [Rotate] [Delete]
deployer      pull, push    blog / api                 created 1 hour ago   [Reveal] [Edit] [Rotate] [Delete]
```

The form is a name, a rights choice of two cards - *Pull only* and *Pull and push* - and a list
of access rows, each a project picker and an optional app picker. A row with no app reads "all
apps in this project", which is what the pattern does.

After creation the dialog shows what to run, with the address already in it:

```
docker login registry.example.com -u ci-github -p <password>
docker pull registry.example.com/hivepaas/shop-api:prod-9f3c1de
```

Two sentences the screen has to carry, because both are consequences an operator cannot see:

> Saving restarts the registry for a few seconds.

> A credential reaches every environment of the apps it names. The environment is part of the
> tag, and the registry can only grant access by repository.

## 7. What this is not

- **Not a way to give someone push access to a project's images casually.** Push means an
  outside party can replace the image a deployment will pull. The form's default is pull only,
  and the push option says this in a line under it.
- **Not per environment**, for the reason in §2.
- **Not zot's API keys.** If a later version wants short-lived tokens, the honest way is an
  issuer HivePaaS controls - a credential with an expiry it enforces by rendering the account
  file without it - not the registry's own keys, which cannot be scoped or revoked centrally.
- **Not for the project-level registries** somebody provisions from the catalogue. Those are
  ordinary apps with their own configuration.

## 8. Testing

- **Rendering.** Golden configurations: one credential with a project, one with a single app,
  one with several rows, one with none. `hivepaas` is in `adminPolicy` in every one of them,
  `**` allows nobody in every one of them, and the account file lists the system line first.
- **Stability.** Rendering twice with the same credentials produces identical bytes, so a save
  that changed nothing does not restart the registry.
- **Validation.** `hivepaas` as a username, a duplicate, an empty access list, a project the
  caller cannot read, an app that is not in the project it is listed under.
- **Deletion.** A deleted credential leaves neither a line nor a rule.
- **End to end**, against a real zot: create a pull-only credential for one project, then with
  that credential pull one of its images, fail to pull another project's, fail to push, and see
  only the permitted repositories in `/v2/_catalog`. Then delete it and watch the same pull be
  refused.

## 9. Later

- Retention per repository pattern, still its own task.
- An expiry on a credential, rendered by leaving it out of the account file once it passes.
- Free-form repository patterns, for images that HivePaaS did not push.
- Per-credential audit: which images a credential pulled, from zot's own logs.
