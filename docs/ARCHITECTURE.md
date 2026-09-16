# Backend architecture

## At a glance

- **One direction only.** `handler → dto → usecase → service → repository → entity → base`. A layer never imports one to its left.
- **Two entry points, one core.** REST (`interface/api`) and agent gRPC (`interface/agent`) both stop at the DTO; everything below is shared and knows about neither.
- **DTO is the wire contract.** Request and response are separate types, named `...Req` / `...Resp`, living in `<name>dto` next to their usecase.
- **Links are objects, not ids.** `entity.ObjectID` in the entity, `basedto.ObjectIDReq` in a request, a `...Resp` object in a response — `{"volume": {"id": "…"}}`, never `{"volumeId": "…"}`.
- **`ModifyRequest()` runs before `Validate()`.** Normalize in the first, check in the second, and never do both in one.
- **Payload goes in `Data`.** A response is `Meta` + `Data`; new fields belong inside `Data`. Lists take `basedto.Paging` in and `basedto.ListMeta` out; everything else uses `basedto.Meta`.
- **Entity becomes response through `Transform<X>()`**, which is also where secrets get masked.
- **Secrets are masked by default.** Revealing one takes a decrypt, a capability and an audit entry.
- **Entity and DTO move together.** Changing an entity field means finding every DTO that carries it.
- **Extract a service the moment a second usecase could want it**, not after the copy exists.
- **Usecases record audit logs** — except those going through `usecase/settings`, which already does.
- **`services/` is the outside world**, with its own errors and its own translation file.

---

## 1. Layers

| Layer | Where | Holds |
|---|---|---|
| handler | `interface/api/handler`, `interface/agent/server` | transport only: bind, call one usecase, render |
| dto | `usecase/<x>uc/<x>dto` | the wire shape, its validation and normalization |
| usecase | `usecase/<x>uc`, `usecaseagent/<x>uc` | one business operation end to end, including its audit log |
| service | `service/<x>service` (+ `<x>serviceimpl`) | logic two or more usecases need |
| repository | `repository` | database access |
| entity | `entity` | what is stored, and how it is stored |
| base | `base`, `basedto`, `hperrors`, `pkg` | vocabulary everything shares |

`services/` (repo root, no `hivepaas_app/`) is separate from all of this — see §6.

A handler does not reach past the DTO, and a usecase does not import another usecase. When two usecases need the same thing, that thing is a service.

## 2. DTO

**One package per usecase, one file per operation.** `usecase/systemsettings/logginguc/loggingdto/` holds `settings_get.go` and `settings_update.go` — read those before writing a new one.

**Naming.** Every type ends in `Req` or `Resp`. Sharing one struct between both directions is reserved for cases where the shapes are genuinely identical and expected to stay that way; it is not the default.

**Object links.** The same relationship is spelled three ways, on purpose:

```go
// entity
Volume ObjectID `json:"volume,omitempty"`

// request
Volume basedto.ObjectIDReq `json:"volume"`        // {"volume": {"id": "…"}}

// response — pick by how much the caller needs
Volume *basedto.NamedObjectResp                    // {id, name}
Volume *settings.BaseSettingResp                   // a whole setting
App    *appdto.AppBaseResp / *appdto.AppResp       // a whole app
```

**The request mirrors the response.** If a response carries `{"volume": {"id": …, "name": …}}`, the request carries `{"volume": {"id": …}}` — the same field name, the same shape, fewer fields. Flattening it to `volumeId` breaks that symmetry and makes validation errors name a field the client never sent.

**`ModifyRequest()` then `Validate()`.** The handler calls them in that order (`BaseHandler.ParseAndValidateRequest`). So:

- `ModifyRequest() error` — implement `basedto.ReqModifier`. Fold input into one spelling: trim, lowercase, drop blanks. Everything downstream, validation included, then reads one form.
- `Validate() hperrors.ValidationErrors` — implement `basedto.ReqValidator`. Check only; never modify.

List query parameters are `[]string`, not a string you split yourself: the decoder already joins repeated parameters and splits on commas, so `?levels=a,b` and `?levels=a&levels=b` both arrive as a slice.

**Responses are `Meta` + `Data`.** New fields go inside `Data`, never beside it:

```go
type GetAppLogHistoryResp struct {
    Meta *basedto.Meta          `json:"meta"`
    Data *AppLogHistoryDataResp `json:"data"`
}
```

**Lists are the one shape with its own meta.** A list request embeds `Paging basedto.Paging` (tagged `json:"-"`, it comes from the query), and its response carries `*basedto.ListMeta` — `Meta` plus the page it describes. Every other endpoint uses `*basedto.Meta`.

**Entity to response goes through `Transform<X>()`.** Name the function for what it produces (`TransformLoggingSettings`, `TransformSettingBase`) and keep the mapping there rather than inline in the usecase, so one place decides what the wire sees — including which fields are hidden.

**Secrets are masked in `Transform`, by default.** An `entity.EncryptedField` never reaches a response as itself:

```go
// copy the stored value into the DTO when it is wanted
func (resp *EndpointResp) CopyPassword(field entity.EncryptedField) error

// otherwise — and this is the default — write the placeholder
resp.Backend.Ingest.Password = basedto.MaskedSecret
```

Masking is what happens unless something deliberately asks otherwise. A new secret field that nobody remembered to mask leaks on the first response.

**Revealing a secret is three gates, not one.** When an endpoint supports `?revealSecrets=true` (`RevealSecrets bool` with `mapstructure:"revealSecrets"`):

1. the server config must allow that secret type (`config.Current().Security.AllowsSecretType`),
2. the caller must hold `base.ResourceCapSecretReveal`,
3. the reveal is recorded as an audit entry.

Decryption belongs in the usecase, not the DTO. `usecase/settings` already implements all of this — an endpoint that goes through it inherits the gates and must not re-implement them.

**swag annotations are one line each.** `@Param` stops at the newline, so it cannot wrap — but the 120-character limit still applies to it. Long prose goes in repeated `@Description` lines instead, each short enough to pass.

**Types that marshal as strings need to say so.** `timeutil.Duration` and `unit.DataSize` write `"30d"` and `"1gb"`, but swag sees the integer underneath. Tag them `swaggertype:"string"` or the generated API contract is wrong.

## 3. Entity

The entity is the stored shape; the DTO is the sent shape. They are tightly coupled by hand, not generated, so **a change to one is a search for the other**. Removing an entity field and leaving it in a response DTO produces a field that is always null — valid Go, broken contract.

Settings entities are JSON blobs, so removing a field needs no migration: it simply stops being read. Renaming or repurposing one does.

## 4. Service

A service is logic **two or more usecases need**. Create it while writing the first usecase if a second plausibly wants it — extracting later means finding every copy that drifted.

Layout is `service/<x>service/` for the interface and types, `service/<x>serviceimpl/` for the implementation, so callers depend on the interface.

**More than two parameters gets a struct.** Beyond `ctx` and `db`, pass a `...Req` and return a `...Resp`:

```go
Apply(ctx context.Context, db database.IDB, req *SettingApplyReq) (*SettingApplyResp, error)
```

Adding a field then costs one line and breaks nobody.

**A service must not import a DTO.** The table in §1 puts `dto` above `service`,
and the reason is not tidiness: a DTO is the dashboard's wire shape and is
expected to move when the UI does. `specservice` is the case that makes this
concrete. It exports app configuration that lives in the Swarm service, which
`appsettingsdto.Transform*` already reads - and it deliberately does not call
them, because a spec is read back by versions of HivePaaS that do not exist yet
and must not change shape for a reason that has nothing to do with it.

Owning a shape that way has a cost, and it is paid in tests rather than in
imports: the Get and Update DTOs carry identical field sets, so anything
readable through them was writable, and a service with its own types gives that
up. `specserviceimpl`'s coverage oracle compares the two field sets in both
directions and fails when they drift. Test code may import anything - a test is
not a layer.

## 5. Usecase

One usecase is one operation, start to finish: load, authorise, act, **record the audit log**.

```go
uc.auditService.Record(ctx, db, &auditservice.Entry{ … })
```

**Exception:** anything going through `usecase/settings` inherits its audit logging and its reveal-secrets authorisation. Do not record a second entry there.

## 6. `services/`

The root `services/` directory is for **systems HivePaaS does not own** (`git`, `aws`, `email`, `im`, `ssh`, `ssl`, `docker`, `cloudflare`, `traefik`) and for **subsystems large enough to own their vocabulary** (`backup`, `logging`).

Rules that make it separate rather than just another folder:

- **It never imports `entity`.** It defines its own types, and the matching `service/` converts at the boundary — credentials cross decrypted.
- **Its own `errors.go`,** built with `hperrors.NewErr(base, "ERR_…")` so every error already carries the code the dashboard maps.
- **Its own translation file,** `pkg/translation/messages/en/errors.<domain>.en.toml`.
- **A model package** (`backupmodel`, `loggingmodel`) holding interfaces and types that depend on no implementation, so a second backend is a new package rather than an edit.
