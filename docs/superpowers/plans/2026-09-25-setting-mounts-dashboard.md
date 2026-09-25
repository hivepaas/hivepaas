# Setting Mounts - Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** An app's Setting Mounts screen, under App settings → Config & Data, where people list, create, edit, enable, disable and delete entries.
- A person sees why an entry is not in the container.
- A person who may not reveal secrets sees the sensitive parts locked, with the reason.

**Architecture:** The screen copies the Config Files slice file for file.

**Data layer:**
- domain entity, API contracts, validator and API class;
- API context and hook;
- query keys, queries and commands.

**Screens:**
- a list route with a table and a menu;
- create and edit routes sharing one form route.

The form picks a source type and a setting of that type, then shows one row per part the type offers:
- a row is ticked to mount the part;
- a path is suggested;
- the mode defaults to `0400` for a sensitive part and `0444` otherwise.

Two pure helpers mirror what the backend decides:
- `suggestPath`;
- `widensGrants`, the (source, sensitive part) pairs a save adds. It blocks the submit, with the reason, for a person who may not reveal.

**Tech Stack:** React 19, TanStack Query, React Hook Form with Zod, shadcn UI, zustand-free (routes, not dialogs).

**Spec:** `docs/superpowers/specs/2026-09-25-setting-mounts-design.md` §10. The routing and kind screens' passthrough state waits for plan 3. Wire format: plan 2 (`docs/superpowers/plans/2026-09-25-setting-mounts-surfaces.md`, Global Constraints).

## Amendment: one way to a file

`docs/superpowers/specs/2026-09-25-setting-mounts-one-way-design.md` (plans 1-3 built) changes this plan. Where the tasks below disagree with this section, this section wins.

- **Wire format.**
  - Files carry `secret` (stored as a Docker secret) and `gated` (takes Reveal Secrets) in place of `sensitive`; so do the parts of `GET .../sources`. `mayMountSensitive` keeps its name.
  - Entries carry `inheritable`, and create and update send it.
  - `ExecuteAppClone` answers with `meta.warning` when it left gated entries out.
- **Two more source types.** `secret` (part `value`) and `config-file` (part `content`). Their picker lists the app's secrets and config files through the app-level queries (`ProjectAppSecretsQueries`, `AppConfigFilesQueries`), which already carry the project's inheritable ones. The labels are "Secret" and "Config file".
- **Marks and locks.** A `gated` part shows the lock and is what `widensGrants` counts and what is locked. A `secret` part that is not gated shows a "stored as secret" mark only.
- **Default mode** is `0400` for a `secret` part and `0444` otherwise.
- **Suggested paths.** A secret's value suggests `/run/secrets/<setting name>`, a config file's content `/etc/app/<setting name>`, both lowercased. Picking a setting replaces a row's path only while it is still the suggestion.
- **Inheritable.** The form has a switch "Previews and clones get this entry", with the line "A preview runs a pull request's code: whatever this entry mounts reaches it." The list marks inheritable entries.
- **Task 5: Secrets and Config Files lose their mount fields.** The `mountIntoFilesystem`, `filePath`, `fileMode`, `fileUid`, `fileGid` fields of both forms (route and dialog), their schemas, `swarmRef` in the entities, contracts, validators and API bodies, and the Mountpoint columns go. Commit: "refactor(apps): secrets and config files are mounted through setting mounts".
- **Task 6: The clone says what it left out.** The execute call parses the response's `meta`, and the clone route shows `meta.warning` with `toast.warning` beside "App clone started". Commit: "feat(apps): the clone says which setting mounts it left out".
- **Task 4** runs after Tasks 5 and 6, and its checks add: a secret mounted at its suggested path; an inheritable switch that survives an edit.

## Global Constraints

- **Repository:** `../hivepaas-dashboard`, on branch `feat/setting-mounts`. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Merge locally, delete the branch, do not push, and stage only the files named here.
- **Wire format** (backend `settingmountdto`):
  - list and get entries: `{id, name, status, updateVer, createdAt, updatedAt, source: {id, name, type, status}, files: [{part, path, uid, gid, mode, sensitive}], state: {reason, mounted: string[]} | null}`;
  - create and update: `{name, source: {id}, files: [{part, path, uid, gid, mode}], updateVer}`, where `mode` is a string such as `"0400"`;
  - status: `PUT .../setting-mounts/{id}/status` with `{status, updateVer}`;
  - `GET .../setting-mounts/sources`: `{data: [{type, parts: [{name, required, sensitive}]}], mayMountSensitive}`.
- **Reasons** (`state.reason`): `entry-disabled`, `key-invalid`, `source-unavailable`, `source-incomplete`, `paths-taken`. An empty reason means mounted, and `state: null` means unknown.
- **Entry key:** `^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$`, never `tls`.
- **Paths:** absolute, and not under `/run/secrets/tls`.
- **Route:** `projects/:id/:env/apps/:appId/setting-mounts` (plus `/create` and `/:settingMountId/edit`). The menu item is "Setting Mounts" under "Config & Data", after "Config Files".
- **Gates:** `npm run lint:ci`, `npm run build`. The dashboard has no test runner, so each task is checked by `tsc` and lint, and the last task by a run against the backend.

## Review Focus

1. **Editing an entry that already mounts a private key, without the permission, must still save a change of path.** Only what is added is gated. *Covered by `widensGrants` in Task 3: it compares against the entry as loaded.*
2. **Changing an entry's source while a sensitive part is ticked is growth.** For someone without the permission, the submit is blocked with the reason, not sent to fail. *Task 3, `widensGrants` counts a new source.*
3. **A disabled entry being edited hands out nothing.** Its grants start empty, so enabling it later is what asks (the backend's rule). *Task 3, `widensGrants(before)` is empty for a disabled entry.*
4. **`state: null` renders as "Unknown", not as mounted.** *Task 2, the state cell.*
5. **Switching the source type drops the parts the old type offered.** A cert's `privateKey` row must not stay ticked under basic auth. *Task 3, the type change resets `files`.*

---

### Task 1: The data layer

**Files (dashboard, `src/application/modules/projects`):**
- Create: `domain/apps/setting-mount/app-setting-mount.entity.ts`, `domain/apps/setting-mount/index.ts`
- Modify: `domain/apps/index.ts` (add `export * from "./setting-mount";`)
- Create: `api/services/project-apps-services/setting-mounts/`:
  - `app-setting-mounts.api.contracts.ts`;
  - `app-setting-mounts.api.validator.ts`;
  - `app-setting-mounts.api.ts`;
  - `index.ts`.
- Modify: `api/services/project-apps-services/index.ts` (`export * from "./setting-mounts";`)
- Modify: `api/api-context/projects.api.context.ts` (validator and `settingMounts: { $: new AppSettingMountsApi(...) }` beside `configFiles`)
- Create: `api/hooks/project-apps/use-app-setting-mounts.api.ts`; modify `api/hooks/project-apps/index.ts`
- Modify: `data/constants/projects.query-keys.ts` (three keys)
- Create: `data/queries/project-apps/app-setting-mounts.queries.ts`; modify `data/queries/project-apps/index.ts`
- Create: `data/commands/project-apps/app-setting-mounts.commands.ts`; modify `data/commands/project-apps/index.ts`
- Modify: `data/commands/project-apps/app-configuration-cache.helpers.ts` (invalidate the two entry keys)

**Interfaces:**
- Produces:
  - `AppSettingMount`, `AppSettingMountFile`, `AppSettingMountState`, `AppSettingMountSource`, `AppSettingMountSourcePart`, `AppSettingMountSources`;
  - `SETTING_MOUNT_REASONS`;
  - `AppSettingMountsQueries.useFindManyPaginated`, `useFindOneById`, `useFindSources`;
  - `AppSettingMountsCommands.useCreateOne`, `useUpdateOne`, `useUpdateStatus`, `useDeleteOne`;
  - request types `AppSettingMounts_*_Req`.

- [ ] **Step 1: Branch.** `cd ../hivepaas-dashboard && git checkout main && git checkout -b feat/setting-mounts`

- [ ] **Step 2: The entity.** `domain/apps/setting-mount/app-setting-mount.entity.ts`:

```ts
import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

/** Why nothing of an entry is mounted; empty when something is. */
export const SETTING_MOUNT_REASONS = {
    "entry-disabled": "The entry is disabled.",
    "key-invalid": "The name is not one an entry may have.",
    "source-unavailable": "The source setting is missing, disabled, or not visible from this app.",
    "source-incomplete": "The source setting has no value yet, such as a certificate not obtained.",
    "paths-taken": "Every path is used by another file of the app.",
} as const;

export type SettingMountReason = keyof typeof SETTING_MOUNT_REASONS;

export interface AppSettingMountFile {
    part: string;
    path: string;
    uid: string;
    gid: string;
    mode: string;
    sensitive: boolean;
}

export interface AppSettingMountState {
    /** Empty when mounted. */
    reason: string;
    mounted: string[];
}

export interface AppSettingMountSourceRef {
    id: string;
    name: string;
    type: string;
    status: string;
}

export interface AppSettingMount {
    id: string;
    /** The entry's key. */
    name: string;
    status: OpenApiConstant<EProjectSecretStatus>;
    updateVer: number;
    source: AppSettingMountSourceRef;
    files: AppSettingMountFile[];
    /** Null when the backend could not read it. */
    state: AppSettingMountState | null;
    createdAt: Date;
    updatedAt: Date | null;
}

export interface AppSettingMountSourcePart {
    name: string;
    required: boolean;
    sensitive: boolean;
}

export interface AppSettingMountSource {
    type: string;
    parts: AppSettingMountSourcePart[];
}

export interface AppSettingMountSources {
    sources: AppSettingMountSource[];
    /** Whether the caller may mount a private key or a password. */
    mayMountSensitive: boolean;
}
```

`index.ts`: `export * from "./app-setting-mount.entity";`

- [ ] **Step 3: Contracts.** `app-setting-mounts.api.contracts.ts`:

```ts
import { type PaginationState, type SortingState } from "@infrastructure/data";
import type { AppSettingMount, AppSettingMountSources } from "~/projects/domain";
import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

type AppScope = {
    projectID: string;
    env: string;
    appID: string;
};

export type AppSettingMountFilePayload = {
    part: string;
    path: string;
    uid: string;
    gid: string;
    mode: string;
};

export type AppSettingMountPayload = {
    name: string;
    sourceID: string;
    files: AppSettingMountFilePayload[];
};

export type AppSettingMounts_FindManyPaginated_Req = ApiRequestBase<
    AppScope & { pagination?: PaginationState; sorting?: SortingState; search?: string }
>;
export type AppSettingMounts_FindManyPaginated_Res = ApiResponsePaginated<AppSettingMount>;

export type AppSettingMounts_FindOneById_Req = ApiRequestBase<AppScope & { settingMountID: string }>;
export type AppSettingMounts_FindOneById_Res = ApiResponseBase<AppSettingMount>;

export type AppSettingMounts_FindSources_Req = ApiRequestBase<AppScope>;
export type AppSettingMounts_FindSources_Res = ApiResponseBase<AppSettingMountSources>;

export type AppSettingMounts_CreateOne_Req = ApiRequestBase<AppScope & AppSettingMountPayload>;
export type AppSettingMounts_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type AppSettingMounts_UpdateOne_Req = ApiRequestBase<
    AppScope & AppSettingMountPayload & { settingMountID: string; updateVer: number }
>;
export type AppSettingMounts_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type AppSettingMounts_UpdateStatus_Req = ApiRequestBase<
    AppScope & { settingMountID: string; updateVer: number; status: EProjectSecretStatus }
>;
export type AppSettingMounts_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type AppSettingMounts_DeleteOne_Req = ApiRequestBase<AppScope & { settingMountID: string }>;
export type AppSettingMounts_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
```

- [ ] **Step 4: Validator.** `app-setting-mounts.api.validator.ts`:

```ts
import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppSettingMounts_CreateOne_Res,
    AppSettingMounts_FindManyPaginated_Res,
    AppSettingMounts_FindOneById_Res,
    AppSettingMounts_FindSources_Res,
} from "~/projects/api/services/project-apps-services";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const AppSettingMountSchema = z.object({
    id: z.string(),
    name: z.string(),
    status: z.string(),
    updateVer: z.number(),
    source: z.object({
        id: z.string(),
        name: z.string().optional().default(""),
        type: z.string().optional().default(""),
        status: z.string().optional().default(""),
    }),
    files: z
        .array(
            z.object({
                part: z.string(),
                path: z.string(),
                uid: z.string().optional().default(""),
                gid: z.string().optional().default(""),
                mode: z.union([z.string(), z.number()]).transform(value => String(value)),
                sensitive: z.boolean().optional().default(false),
            }),
        )
        .nullable()
        .transform(value => value ?? []),
    state: z
        .object({
            reason: z.string().optional().default(""),
            mounted: z
                .array(z.string())
                .nullable()
                .optional()
                .transform(value => value ?? []),
        })
        .nullable()
        .optional()
        .default(null),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AppSettingMountSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AppSettingMountSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const FindSourcesSchema = z.object({
    data: z
        .array(
            z.object({
                type: z.string(),
                parts: z.array(
                    z.object({
                        name: z.string(),
                        required: z.boolean(),
                        sensitive: z.boolean(),
                    }),
                ),
            }),
        )
        .nullable()
        .transform(value => value ?? []),
    mayMountSensitive: z.boolean(),
    meta: BaseMetaApiSchema.nullable().optional().default(null),
});

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppSettingMountsApiValidator {
    findManyPaginated = (response: AxiosResponse): AppSettingMounts_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): AppSettingMounts_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    findSources = (response: AxiosResponse): AppSettingMounts_FindSources_Res => {
        const { data, mayMountSensitive, meta } = parseApiResponse({ response, schema: FindSourcesSchema });
        return { data: { sources: data, mayMountSensitive }, meta };
    };

    createOne = (response: AxiosResponse): AppSettingMounts_CreateOne_Res => {
        return parseApiResponse({ response, schema: CreateOneSchema });
    };
}
```

Check how `ApiResponseBase` types `meta`: `grep -n "ApiResponseBase" -A6 src/infrastructure/api/*.ts`. If `meta` is required and non-nullable, keep what `findOneById` returns (`meta`) and match that shape in `findSources`. The status `z.string()` is typed as `OpenApiConstant<...>` in the entity, as config files do.

- [ ] **Step 5: API class.** `app-setting-mounts.api.ts`:

```ts
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppSettingMountPayload,
    AppSettingMountsApiValidator,
    AppSettingMounts_CreateOne_Req,
    AppSettingMounts_CreateOne_Res,
    AppSettingMounts_DeleteOne_Req,
    AppSettingMounts_DeleteOne_Res,
    AppSettingMounts_FindManyPaginated_Req,
    AppSettingMounts_FindManyPaginated_Res,
    AppSettingMounts_FindOneById_Req,
    AppSettingMounts_FindOneById_Res,
    AppSettingMounts_FindSources_Req,
    AppSettingMounts_FindSources_Res,
    AppSettingMounts_UpdateOne_Req,
    AppSettingMounts_UpdateOne_Res,
    AppSettingMounts_UpdateStatus_Req,
    AppSettingMounts_UpdateStatus_Res,
} from "~/projects/api/services/project-apps-services";

import { BaseApi, parseApiError } from "@infrastructure/api";

function basePath(projectID: string, env: string, appID: string): string {
    return `/projects/${projectID}/${env}/apps/${appID}/setting-mounts`;
}

function toBody({ name, sourceID, files }: AppSettingMountPayload) {
    return {
        name,
        source: { id: sourceID },
        files: files.map(file => ({
            part: file.part,
            path: file.path,
            uid: file.uid,
            gid: file.gid,
            mode: file.mode,
        })),
    };
}

export class AppSettingMountsApi extends BaseApi {
    public constructor(private readonly validator: AppSettingMountsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AppSettingMounts_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_FindManyPaginated_Res, Error>> {
        const { projectID, env, appID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get(basePath(projectID, env, appID), { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: AppSettingMounts_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_FindOneById_Res, Error>> {
        const { projectID, env, appID, settingMountID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${basePath(projectID, env, appID)}/${settingMountID}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findSources(
        request: AppSettingMounts_FindSources_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_FindSources_Res, Error>> {
        const { projectID, env, appID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${basePath(projectID, env, appID)}/sources`, { signal })).pipe(
                map(this.validator.findSources),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: AppSettingMounts_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_CreateOne_Res, Error>> {
        const { projectID, env, appID, ...payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(basePath(projectID, env, appID), toBody(payload), { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: AppSettingMounts_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_UpdateOne_Res, Error>> {
        const { projectID, env, appID, settingMountID, updateVer, ...payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `${basePath(projectID, env, appID)}/${settingMountID}`,
                    { ...toBody(payload), updateVer },
                    { signal },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: AppSettingMounts_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSettingMounts_UpdateStatus_Res, Error>> {
        const { projectID, env, appID, settingMountID, updateVer, status } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `${basePath(projectID, env, appID)}/${settingMountID}/status`,
                    { status, updateVer },
                    { signal },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: AppSettingMounts_DeleteOne_Req): Promise<Result<AppSettingMounts_DeleteOne_Res, Error>> {
        const { projectID, env, appID, settingMountID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${basePath(projectID, env, appID)}/${settingMountID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
```

`index.ts` exports the three files, as `config-files/index.ts` does.

- [ ] **Step 6: Context, hook, keys, queries, commands.**

**Context.** In `projects.api.context.ts`:
- `const appSettingMountsApiValidator = new AppSettingMountsApiValidator();` beside `appConfigFilesApiValidator`;
- `settingMounts: { $: new AppSettingMountsApi(appSettingMountsApiValidator) },` after `configFiles`;
- add both names to the import from `~/projects/api/services`, beside `AppConfigFilesApi`.

**Query keys.** In `projects.query-keys.ts`, after the config-files keys:

```ts
    "projects.apps.setting-mounts.$.find-many-paginated": "projects.apps.setting-mounts.$.find-many-paginated",
    "projects.apps.setting-mounts.$.find-one-by-id": "projects.apps.setting-mounts.$.find-one-by-id",
    "projects.apps.setting-mounts.$.find-sources": "projects.apps.setting-mounts.$.find-sources",
```

**Hook.** `use-app-setting-mounts.api.ts` follows `use-app-config-files.api.ts` exactly, over `api.projects.apps.settingMounts.$`:
- `queries`: `findManyPaginated`, `findOneById`, `findSources`, each `match(result, { Ok: _ => _, Err: error => { throw error; } })`;
- `mutations`: `createOne`, `updateOne`, `updateStatus`, `deleteOne`, each calling `notifyError({ message: "Failed to <verb> setting mount", error })` before throwing;
- no `helpers`.

The hook is exported as `useAppSettingMountsApi`. Add `export * from "./use-app-setting-mounts.api";` to the hooks index.

**Queries.** `app-setting-mounts.queries.ts`:

```ts
import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppSettingMountsApi } from "~/projects/api";
import type {
    AppSettingMounts_FindManyPaginated_Req,
    AppSettingMounts_FindManyPaginated_Res,
    AppSettingMounts_FindOneById_Req,
    AppSettingMounts_FindOneById_Res,
    AppSettingMounts_FindSources_Req,
    AppSettingMounts_FindSources_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppSettingMounts_FindManyPaginated_Req["data"];
type FindManyPaginatedOptions = Omit<
    UseQueryOptions<AppSettingMounts_FindManyPaginated_Res>,
    "queryKey" | "queryFn"
>;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppSettingMountsApi();

    return useQuery({
        queryKey: [QK["projects.apps.setting-mounts.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = AppSettingMounts_FindOneById_Req["data"];
type FindOneByIdOptions = Omit<UseQueryOptions<AppSettingMounts_FindOneById_Res>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppSettingMountsApi();

    return useQuery({
        queryKey: [QK["projects.apps.setting-mounts.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

type FindSourcesReq = AppSettingMounts_FindSources_Req["data"];
type FindSourcesOptions = Omit<UseQueryOptions<AppSettingMounts_FindSources_Res>, "queryKey" | "queryFn">;

function useFindSources(request: FindSourcesReq, options: FindSourcesOptions = {}) {
    const { queries } = useAppSettingMountsApi();

    return useQuery({
        queryKey: [QK["projects.apps.setting-mounts.$.find-sources"], request],
        queryFn: ({ signal }) => queries.findSources(request, signal),
        ...options,
    });
}

export const AppSettingMountsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    useFindSources,
});
```

**Commands.** `app-setting-mounts.commands.ts` follows `app-config-files.commands.ts`:
- four commands, `useCreateOne`, `useUpdateOne`, `useUpdateStatus` and `useDeleteOne`, each taking `mutationFn: mutations.<name>`;
- each calls `invalidateSingleAppConfigurationQueries(queryClient, { projectID: request.projectID, appID: request.appID })` on success, then the caller's `onSuccess`;
- exported as `AppSettingMountsCommands`.

**Cache.** In `app-configuration-cache.helpers.ts`, after the config-files keys, invalidate `projects.apps.setting-mounts.$.find-many-paginated` and `projects.apps.setting-mounts.$.find-one-by-id` the same way.

**Indexes.** Add each new file to the `index.ts` of its folder.

- [ ] **Step 7: Check.** Run `npx tsc --noEmit -p .` and `npm run lint:ci`. Both should be clean.

- [ ] **Step 8: Commit**

```bash
git add src/application/modules/projects/domain src/application/modules/projects/api src/application/modules/projects/data
git commit -m "feat(apps): the setting mounts data layer

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: The list screen, the menu, and the way there

**Files (dashboard):**
- Modify: `src/application/shared/constants/route.constants.ts`. Add `settingMounts` after `configFiles`, with `$pattern`/`$route` for the list, `create`, and `edit` (`:settingMountId/edit`), as `configFiles` has.
- Create: `src/application/modules/projects/module-shared/definitions/tables/app-setting-mounts/`:
  - `app-setting-mounts-table.defs.tsx`;
  - `index.ts`;
  - `building-blocks/edit-cell.com.tsx`, `building-blocks/menu-cell.com.tsx`, `building-blocks/state-cell.com.tsx`, `building-blocks/index.ts`.
- Create: `src/application/modules/projects/routes/single-project/single-app/configuration/setting-mounts/`:
  - `route/app-setting-mounts.route.com.tsx`;
  - `route/index.ts`;
  - `index.ts`.
- Modify: `routes/single-project/single-app/configuration/index.ts` (`export * from "./setting-mounts";`), `routes/index.ts`, `projects.module.ts`, `projects.router.tsx` (the list route; create and edit come in Task 3).
- Modify: `layouts/single-app/configuration-layout/single-app-configuration-layout.com.tsx`: the item `{ label: "Setting Mounts", icon: FileKey, route: ...settingMounts.$route(projectId, env, appId) }` after "Config Files". Import `FileKey` from `lucide-react`.
- Modify: `layouts/single-app/header/single-app-header.com.tsx` (the active prefix, after `configFiles`).

**Interfaces:**
- Consumes: Task 1's queries, commands and entity.
- Produces: `ROUTE.projects.single.apps.single.configuration.settingMounts.{$route,create.$route,edit.$route}` and `AppSettingMountsTableDefs.columns(projectId, env, appId)`.

- [ ] **Step 1: Route constants:**

```ts
                        settingMounts: {
                            $pattern: "projects/:id/:env/apps/:appId/setting-mounts",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/setting-mounts/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/setting-mounts/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/setting-mounts/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/:env/apps/:appId/setting-mounts/:settingMountId/edit",
                                $route: (id: string, env: string, appId: string, settingMountId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/setting-mounts/${settingMountId}/edit/`,
                            },
                        },
```

- [ ] **Step 2: The state cell.** `building-blocks/state-cell.com.tsx`:

```tsx
import React from "react";

import { Badge } from "@components/ui/badge";
import { SETTING_MOUNT_REASONS, type AppSettingMount, type SettingMountReason } from "~/projects/domain";

function reasonText(reason: string): string {
    return SETTING_MOUNT_REASONS[reason as SettingMountReason] ?? reason;
}

function View({ settingMount }: Props) {
    const { state } = settingMount;

    if (!state) {
        return <Badge variant="outline">Unknown</Badge>;
    }

    if (state.reason) {
        return (
            <div className="flex flex-col gap-1">
                <Badge className="w-fit bg-amber-500 text-white">Not mounted</Badge>
                <span className="text-xs text-muted-foreground">{reasonText(state.reason)}</span>
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-1">
            <Badge className="w-fit bg-green-600 text-white">Mounted</Badge>
            {state.mounted.map(path => (
                <code
                    key={path}
                    className="break-all text-xs text-muted-foreground"
                >
                    {path}
                </code>
            ))}
        </div>
    );
}

interface Props {
    settingMount: AppSettingMount;
}

export const StateCell = React.memo(View);
```

Check that `Badge` takes `variant="outline"` (`grep -n "variant" src/components/ui/badge.tsx`). If it does not, use a `className` with a border.

- [ ] **Step 3: Edit cell and menu.** Build `edit-cell.com.tsx` from config-files' edit cell, navigating to `settingMounts.edit.$route(projectId, env, appId, settingMount.id)`.

Build `menu-cell.com.tsx` from config-files' menu cell. Drop the download item, keep "Remove", and add a status toggle above it:

```tsx
    const { mutate: updateStatus, isPending: isUpdatingStatus } = AppSettingMountsCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success(isActive ? "Setting mount disabled" : "Setting mount enabled");
            setOpen(false);
        },
    });
    const isActive = settingMount.status === EProjectSecretStatus.Active;
```

```tsx
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                        triggerClassName="w-full"
                    >
                        {({ isDenied }) => (
                            <Button
                                className="justify-start py-1.5 w-full"
                                variant="ghost"
                                disabled={isDenied || isUpdatingStatus}
                                onClick={() => {
                                    updateStatus({
                                        projectID: projectId,
                                        env,
                                        appID: appId,
                                        settingMountID: settingMount.id,
                                        updateVer: settingMount.updateVer,
                                        status: isActive ? EProjectSecretStatus.Disabled : EProjectSecretStatus.Active,
                                    });
                                }}
                            >
                                {isActive ? <PowerOffIcon className="mr-2 size-4" /> : <PowerIcon className="mr-2 size-4" />}
                                {isActive ? "Disable" : "Enable"}
                            </Button>
                        )}
                    </PermissionTooltipAction>
```

Import `EProjectSecretStatus` from `~/projects/module-shared/enums` and `PowerIcon`, `PowerOffIcon` from `lucide-react`. Enabling an entry with a private key goes through the backend's Reveal Secrets gate, and a refusal comes back as an error toast through `notifyError`.

- [ ] **Step 4: Table definitions.** `app-setting-mounts-table.defs.tsx`:

```tsx
import type { ColumnDef } from "@tanstack/react-table";
import { LockIcon } from "lucide-react";
import type { AppSettingMount } from "~/projects/domain";
import { ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { EditCell, MenuCell, StateCell } from "./building-blocks";

const SOURCE_TYPE_LABELS: Record<string, string> = {
    "ssl-cert": "SSL certificate",
    "ssh-key": "SSH key",
    "basic-auth": "Basic auth",
};

function createColumns(projectId: string, env: string, appId: string): ColumnDef<AppSettingMount>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <EditCell
                    projectId={projectId}
                    env={env}
                    appId={appId}
                    settingMount={original}
                />
            ),
            meta: { align: "center", titleAlign: "center" },
        },
        {
            accessorKey: "name",
            header: "Name",
            cell: ({ row: { original } }) => <div className="break-all">{original.name}</div>,
        },
        {
            header: "Source",
            cell: ({ row: { original } }) => (
                <div className="flex flex-col">
                    <span className="break-all">{original.source.name || original.source.id}</span>
                    <span className="text-xs text-muted-foreground">
                        {SOURCE_TYPE_LABELS[original.source.type] ?? original.source.type}
                    </span>
                </div>
            ),
        },
        {
            header: "Files",
            cell: ({ row: { original } }) => (
                <div className="flex flex-col gap-1">
                    {original.files.map(file => (
                        <div
                            key={file.part}
                            className="flex items-center gap-1 text-xs"
                        >
                            {file.sensitive && <LockIcon className="size-3 text-amber-600" />}
                            <span className="text-muted-foreground">{file.part}</span>
                            <code className="break-all">{file.path}</code>
                        </div>
                    ))}
                </div>
            ),
        },
        {
            header: "Status",
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center">
                    <ProjectSecretStatusBadge status={original.status} />
                </div>
            ),
            meta: { align: "center", titleAlign: "center" },
        },
        {
            header: "In the container",
            cell: ({ row: { original } }) => <StateCell settingMount={original} />,
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <MenuCell
                    projectId={projectId}
                    env={env}
                    appId={appId}
                    settingMount={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const AppSettingMountsTableDefs = Object.freeze({
    columns: createColumns,
    sourceTypeLabels: SOURCE_TYPE_LABELS,
});
```

- [ ] **Step 5: The list route.** `route/app-setting-mounts.route.com.tsx` is `AppConfigFilesRoute`, changed as follows:
- `AppSettingMountsQueries.useFindManyPaginated`;
- `AppSettingMountsTableDefs`;
- the button reads "New Setting Mount" and goes to `settingMounts.create.$route`;
- a line above the table, explaining the screen:

```tsx
            <p className="text-sm text-muted-foreground">
                Mount parts of another setting - a certificate and its key, an SSH key, a basic auth pair as htpasswd -
                as files in this app&apos;s containers. The files follow the setting: a renewed certificate reaches
                the container on its own.
            </p>
```

Register `AppSettingMountsRoute` in:
- `routes/index.ts`;
- `projects.module.ts`, beside `AppConfigFilesRoute`;
- `projects.router.tsx`, after the config-files routes:

```tsx
                        {
                            path: ROUTE.projects.single.apps.single.configuration.settingMounts.$pattern,
                            lazy: async () => {
                                const { AppSettingMountsRoute } = await getLazyComponents();

                                return { Component: AppSettingMountsRoute };
                            },
                        },
```

Also add the menu item and the header prefix.

- [ ] **Step 6: Check.** `npm run lint:ci` and `npm run build` should be clean.

- [ ] **Step 7: Commit**

```bash
git add src/application/shared/constants/route.constants.ts src/application/modules/projects
git commit -m "feat(apps): the setting mounts list, with what is in the container and why not

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The form, create and edit

**Files (dashboard, `src/application/modules/projects/routes/single-project/single-app/configuration/setting-mounts/`):**
- Create:
  - `schemas/app-setting-mount.form.schema.ts`, `schemas/index.ts`;
  - `utils/setting-mount-form.utils.ts`, `utils/index.ts` (`suggestPath`, `defaultMode`, `widensGrants`);
  - `form/app-setting-mount.form.com.tsx`, `form/source-picker.com.tsx`, `form/index.ts`;
  - `form-route/app-setting-mount-form-route.com.tsx`, `form-route/index.ts`;
  - `create/app-setting-mount-create.route.com.tsx`, `create/index.ts`;
  - `edit/app-setting-mount-edit.route.com.tsx`, `edit/index.ts`.
- Modify: the folder's `index.ts`, `routes/index.ts`, `projects.module.ts`, `projects.router.tsx` (the create and edit routes).

**Interfaces:**
- Consumes: Tasks 1 and 2; `ProjectSslCertQueries`, `ProjectSSHKeyQueries`, `ProjectBasicAuthQueries` (`useFindManyPaginated({ projectID, env, search })`), all exported from `~/projects/data/queries`.
- Produces:
  - `AppSettingMountFormSchema`, `AppSettingMountFormInput`, `AppSettingMountFormOutput`;
  - `suggestPath(type, part): string`;
  - `defaultMode(sensitive): string`;
  - `widensGrants(before: Grant[], after: Grant[]): Grant[]`.

- [ ] **Step 1: The helpers.** `utils/setting-mount-form.utils.ts`:

```ts
/** Where a part's file goes unless the person says otherwise. */
const SUGGESTED_PATHS: Record<string, Record<string, string>> = {
    "ssl-cert": {
        certificate: "/etc/app/tls/cert.pem",
        privateKey: "/etc/app/tls/key.pem",
        caCertificate: "/etc/app/tls/ca.pem",
    },
    "ssh-key": {
        privateKey: "/etc/app/ssh/id_key",
        publicKey: "/etc/app/ssh/id_key.pub",
    },
    "basic-auth": {
        username: "/etc/app/auth/username",
        password: "/etc/app/auth/password",
        htpasswd: "/etc/app/auth/htpasswd",
    },
};

export function suggestPath(type: string, part: string): string {
    return SUGGESTED_PATHS[type]?.[part] ?? `/etc/app/${part}`;
}

/** A sensitive file is readable by its owner only; any other, by everyone. */
export function defaultMode(sensitive: boolean): string {
    return sensitive ? "0400" : "0444";
}

export interface Grant {
    source: string;
    part: string;
}

/**
 * The sensitive pairs after hands out that before did not: what the backend
 * asks the Reveal Secrets permission for (§7 of the setting mounts design).
 * A disabled entry hands out nothing, so its before is empty.
 */
export function widensGrants(before: Grant[], after: Grant[]): Grant[] {
    return after.filter(grant => !before.some(held => held.source === grant.source && held.part === grant.part));
}
```

- [ ] **Step 2: The schema.** `schemas/app-setting-mount.form.schema.ts`:

```ts
import { z } from "zod";

export const SETTING_MOUNT_KEY_PATTERN = /^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$/;
export const SETTING_MOUNT_TLS_DIR = "/run/secrets/tls";

const FileRowSchema = z.object({
    part: z.string(),
    sensitive: z.boolean(),
    enabled: z.boolean(),
    path: z.string(),
    mode: z.string(),
    uid: z.string(),
    gid: z.string(),
});

export const AppSettingMountFormSchema = z
    .object({
        name: z
            .string()
            .trim()
            .regex(SETTING_MOUNT_KEY_PATTERN, "Lowercase letters, digits and hyphens, at most 20")
            .refine(value => value !== "tls", "'tls' is reserved for TLS passthrough"),
        sourceType: z.string().min(1, "Choose what to mount from"),
        source: z.object({ id: z.string(), name: z.string() }).nullable(),
        files: z.array(FileRowSchema),
    })
    .superRefine((value, ctx) => {
        if (!value.source?.id) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Choose a setting", path: ["source"] });
        }
        const enabled = value.files.filter(file => file.enabled);
        if (enabled.length === 0) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Mount at least one part", path: ["files"] });
        }
        const seen = new Set<string>();
        value.files.forEach((file, index) => {
            if (!file.enabled) {
                return;
            }
            const path = file.path.trim();
            const message = !path.startsWith("/")
                ? "Use an absolute path"
                : path === SETTING_MOUNT_TLS_DIR || path.startsWith(`${SETTING_MOUNT_TLS_DIR}/`)
                  ? `${SETTING_MOUNT_TLS_DIR} is reserved for TLS passthrough`
                  : seen.has(path)
                    ? "Another part of this entry has this path"
                    : "";
            if (message) {
                ctx.addIssue({ code: z.ZodIssueCode.custom, message, path: ["files", index, "path"] });
            }
            seen.add(path);
        });
    });

export type AppSettingMountFormInput = z.input<typeof AppSettingMountFormSchema>;
export type AppSettingMountFormOutput = z.output<typeof AppSettingMountFormSchema>;
```

- [ ] **Step 3: The source picker.** `form/source-picker.com.tsx` is a `Combobox` over the settings of the chosen type, fed by the matching env-level query. Only the query for the selected type is enabled:

```tsx
import React, { useMemo, useState } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectBasicAuthQueries, ProjectSSHKeyQueries, ProjectSslCertQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { Combobox } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

function View({ sourceType, value, onChange, invalid, disabled }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    const [search, setSearch] = useState("");
    const request = { projectID: projectId, env, search };

    const certs = ProjectSslCertQueries.useFindManyPaginated(request, { enabled: sourceType === "ssl-cert" });
    const keys = ProjectSSHKeyQueries.useFindManyPaginated(request, { enabled: sourceType === "ssh-key" });
    const auths = ProjectBasicAuthQueries.useFindManyPaginated(request, { enabled: sourceType === "basic-auth" });
    const active = sourceType === "ssl-cert" ? certs : sourceType === "ssh-key" ? keys : auths;
    const { data: { data: settings } = DEFAULT_PAGINATED_DATA } = active;

    const options = useMemo(() => {
        const list = settings.map(setting => ({
            value: { id: setting.id, name: setting.name },
            label: setting.name,
        }));
        if (value?.id && !list.some(item => item.value.id === value.id)) {
            list.unshift({ value: { id: value.id, name: value.name || value.id }, label: value.name || value.id });
        }
        return list;
    }, [settings, value]);

    return (
        <Combobox
            options={options}
            value={value?.id ?? null}
            onChange={(_, option) => {
                onChange(option?.id ? { id: option.id, name: option.name } : null);
            }}
            onSearch={setSearch}
            placeholder="Select a setting"
            searchable
            closeOnSelect
            emptyText="No setting of this type is available to this app"
            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
            valueKey="id"
            aria-invalid={invalid}
            loading={active.isFetching}
            onRefresh={() => void active.refetch()}
            isRefreshing={active.isRefetching}
            disabled={disabled}
        />
    );
}

interface Props {
    sourceType: string;
    value: { id: string; name: string } | null;
    onChange: (value: { id: string; name: string } | null) => void;
    invalid?: boolean;
    disabled?: boolean;
}

export const SourcePicker = React.memo(View);
```

Check the paginated item types: `SettingSslCert`, `SettingSSHKey` and `SettingBasicAuth` should all have `id` and `name`. Check also that the three queries accept `{ enabled }` as their second argument (`Omit<UseQueryOptions, ...>`). They do in `project-basic-auth.queries.ts`.

- [ ] **Step 4: The form.** `form/app-setting-mount.form.com.tsx`:
- **Layout** follows `CreateOrEditAppConfigFileForm`: `InfoBlock` rows at `titleWidth={240}`, `FormActionBar` at the bottom, a `fieldset disabled={readOnly}`.
- **Name:** an `Input` with the hint "Lowercase letters, digits and hyphens. Names the Docker objects that hold the files."
- **Source type:** `Tabs` with one trigger per `sources[].type`, labeled with `AppSettingMountsTableDefs.sourceTypeLabels`. Changing it resets `source` to `null` and `files` to that type's parts, all unticked, each with `suggestPath(type, part)`, `defaultMode(part.sensitive)` and empty uid and gid.
- **Setting:** `SourcePicker`.
- **Files:** one row per part:
  - a `Checkbox` "Mount";
  - the part's name, and a `LockIcon` "Sensitive" badge when sensitive;
  - the path `Input`;
  - mode, uid and gid `Input`s at `max-w-[120px]`, with placeholders `mode`, `uid`, `gid`;
  - `FieldError` for `files.{index}.path`.
- **Locking.** For a person who may not reveal (`mayMountSensitive === false`), a sensitive part not in `lockedAllowed` is disabled, with the line "Mounting this reveals it to whoever runs the app: it takes the Can Reveal Secrets permission, and Return Secrets Via API turned on in System → HivePaaS → Security."
- **Submit:** compute `widensGrants(initialGrants, grantsOf(values))`, with `grantsOf` taking the ticked sensitive rows as `{ source: values.source.id, part }`. If that is non-empty and `mayMountSensitive` is false, set a root error with that same line and do not call `onSubmit`.

The component's props:

```tsx
interface Props {
    sources: AppSettingMountSource[];
    mayMountSensitive: boolean;
    isPending: boolean;
    isEditMode: boolean;
    readOnly?: boolean;
    initialValues?: AppSettingMountFormInput;
    /** The grants the entry holds as loaded: none for a new or a disabled entry. */
    initialGrants: Grant[];
    onSubmit: (values: AppSettingMountFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    onClose?: () => void;
    stickyActions?: boolean;
}
```

The files are a `useFieldArray({ control, name: "files" })`. The type reset is `replace(rowsFor(type))`, where:

```ts
function rowsFor(sources: AppSettingMountSource[], type: string): AppSettingMountFormInput["files"] {
    const source = sources.find(item => item.type === type);
    return (source?.parts ?? []).map(part => ({
        part: part.name,
        sensitive: part.sensitive,
        enabled: false,
        path: suggestPath(type, part.name),
        mode: defaultMode(part.sensitive),
        uid: "",
        gid: "",
    }));
}
```

A sensitive row is locked when:

```ts
const locked = !mayMountSensitive && row.sensitive &&
    !initialGrants.some(grant => grant.source === currentSourceId && grant.part === row.part);
```

`currentSourceId` is `useWatch({ control, name: "source" })?.id`.

- [ ] **Step 5: The form route.** `form-route/app-setting-mount-form-route.com.tsx` follows `AppConfigFileFormRoute`:
- It loads `AppSettingMountsQueries.useFindSources({ projectID, env, appID })`, and in edit mode `useFindOneById`.
- It shows `AppLoader` while either loads.
- It builds `initialValues` from the entry, with a row for every part of its type:
  - ticked, with the entry's file values, where the entry has that part;
  - unticked, with suggestions, where it does not.
- `initialGrants`:
  - the entry's sensitive files as `{ source: entry.source.id, part }`, only when `entry.status === "active"`;
  - `[]` for a new entry.
- On submit, it maps the ticked rows to `files` and calls `useCreateOne` or `useUpdateOne`. The update carries `updateVer`, `name`, `sourceID` and `files`.
- On success it toasts and navigates back to the list, as config files do.

The submit mapping:

```ts
function toPayload(values: AppSettingMountFormOutput) {
    return {
        name: values.name,
        sourceID: values.source?.id ?? "",
        files: values.files
            .filter(file => file.enabled)
            .map(file => ({ part: file.part, path: file.path.trim(), uid: file.uid, gid: file.gid, mode: file.mode })),
    };
}
```

The create and edit routes are `AppConfigFileCreateRoute` and `AppConfigFileEditRoute` with `settingMountId` as the param.

Register them in:
- `routes/index.ts`;
- `projects.module.ts`;
- `projects.router.tsx`, with the `settingMounts.create.$pattern` and `settingMounts.edit.$pattern` entries after the list route.

- [ ] **Step 6: Check.** `npm run lint:ci` and `npm run build` should be clean.

- [ ] **Step 7: Commit**

```bash
git add src/application/modules/projects
git commit -m "feat(apps): create and edit setting mounts, with sensitive parts locked for who may not reveal them

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: Against the backend, and merge

- [ ] **Step 1: Run it.** Start the backend of this repository on a port of your own (`HP_*` env variables as `docs/` and memory say; the user's runs on 10000). Then run the dashboard's dev server pointed at it, and check:
  1. An app's "Setting Mounts" item opens the list, with the explanation line.
  2. "New Setting Mount" → SSL certificate → a certificate → tick `certificate` and `privateKey` → Save. The list shows the entry. If the app has a service, "In the container" says Mounted with both paths; if it has none, "Unknown" or the reason.
  3. Switching the source type in the form resets the rows.
  4. Disable, then enable, from the menu. The state follows.
  5. As a member without Can Reveal Secrets, the `privateKey` row is disabled with the reason, and saving `certificate` alone works.

  If the backend cannot be run here, say so, skip to the gates, and list these checks for the user.
- [ ] **Step 2: Gates.** `npm run lint:ci`, `npm run build`.
- [ ] **Step 3: Merge.** `git checkout main && git merge --no-ff feat/setting-mounts -m "Merge branch 'feat/setting-mounts'" && git branch -d feat/setting-mounts`
