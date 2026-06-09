## API Service Guidelines

This guideline documents the standard process to build a complete API surface for any entity in LocalPaaS Dashboard. It covers service and validator implementation, wiring into the API context, React hooks, data queries, and commands. Follow alongside `docs/DEVELOPMENT_GUIDELINES.md` for cross-cutting rules (imports, Zod, Result pattern, query keys, etc.).

### Scope and Outcomes

- **Service**: Type-safe HTTP client methods using `BaseApi` + `Result` pattern.
- **Validator**: Zod schemas using existing meta schemas (e.g., `PagingMetaApiSchema`).
- **Context**: Register the service in the module’s API context factory.
- **Hooks**: React hooks providing `queries` and `mutations` that unwrap `Result` and surface errors with notifications.
- **Queries**: TanStack Query consumers for read operations with proper query keys and placeholder data.
- **Commands**: TanStack Mutation consumers for write operations with cache invalidation.
- **Exports**: Update all related `index.ts` barrels.
- **Mappers** (optional): When request/response JSON differs from domain types, use `<entity>.api.mapper.ts` next to the service (see below).

---

## Step-by-step Checklist

### 1) Define Contracts (types)

- Create a contracts file under the module service folder. Include:
    - Entity shape (domain-aligned fields)
    - Request/response types based on `ApiRequestBase`/`ApiResponseBase`/`ApiResponsePaginated`
    - Support `pagination`/`sorting`/`search` in the request `data` shape when the endpoint needs them

Example (concise):

```ts
import type { PaginationState, SortingState } from "@infrastructure/data";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export interface MyEntity {
    id: string;
    name: string;
}

export type MyEntities_FindMany_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type MyEntities_FindMany_Res = ApiResponsePaginated<MyEntity>;

export type MyEntities_CreateOne_Req = ApiRequestBase<{ payload: { name: string } }>;
export type MyEntities_CreateOne_Res = ApiResponseBase<{ type: "success" }>;
```

### 2) Implement Validator (Zod)

- Create a validator using Zod and `parseApiResponse`.
- Always use shared meta schemas (e.g., `PagingMetaApiSchema`).
- Use `z.coerce.date()` for date fields and avoid string date types.

Example:

```ts
import { type AxiosResponse } from "axios";
import { z } from "zod";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const MyEntitySchema = z.object({ id: z.string(), name: z.string() });
const FindManySchema = z.object({ data: z.array(MyEntitySchema), meta: PagingMetaApiSchema });

export class MyEntitiesApiValidator {
    findMany = (response: AxiosResponse) => {
        const { data, meta } = parseApiResponse({ response, schema: FindManySchema });
        return { data, meta };
    };
}
```

### 2.5) Mappers (domain ↔ wire)

Use when the HTTP payload or parsed response shape **differs** from your domain model (e.g. `null` vs empty array, date string formatting, flattened vs nested objects).

- **Wire types** live in **contracts** (e.g. `MyEntityUpdateBody` for a `PUT` body).
- **Mapping** lives in **`<entity>.api.mapper.ts`**, colocated with the service—not in TanStack queries/commands or generic “data” helpers.
- **Pattern**: one small class per operation with explicit methods:
    - **`toApi`**: domain → wire (request bodies, query serialization if needed).
    - **`toDomain`**: validated API output → domain (when you normalize beyond what Zod already gives you).
- **Facade**: export a single **`MyEntityApiMapper`** with `readonly` nested mappers (e.g. `readonly updateOne = new UpdateOneMapper()`), matching nearby module style.
- **Service**: inject the mapper in the service constructor (alongside the validator) and call `this.mapper.updateOne.toApi(payload)` before `put`/`post`.

This keeps the **data layer** free of business shaping: queries/commands call the API hook; the **service** owns HTTP + validation + domain/wire mapping at the boundary.

### 3) Implement Service (BaseApi + Result)

- Create a service class extending `BaseApi` that:
    - Uses `this.client.v1` for HTTP
    - Builds query params via `this.queryBuilder.getInstance()`
    - Returns `Result` (`Ok`/`Err`) and transforms via validator

Example methods:

```ts
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class MyEntitiesApi extends BaseApi {
    constructor(private readonly validator: MyEntitiesApiValidator) {
        super();
    }

    async findMany(
        req: MyEntities_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<MyEntities_FindMany_Res, Error>> {
        const { pagination, sorting, search } = req.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/my-entities", {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findMany),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
```

### 4) Export Service (Barrel)

- Add a local `index.ts` in the service folder to export contracts, validator, and service.
- Update the module-level aggregator (e.g., `projects-services/index.ts`).

### 5) Wire into API Context

- Import your service and validator in the module API context factory (e.g., `projects.api.context.ts`).
- Instantiate the validator (and **mapper** if the service uses one) and add your service into the returned `api` object under the appropriate subtree.

Example (excerpt):

```ts
import { MyEntitiesApi, MyEntitiesApiMapper, MyEntitiesApiValidator } from "~/<module>/api/services";

const myEntitiesValidator = new MyEntitiesApiValidator();
const myEntitiesMapper = new MyEntitiesApiMapper();

return {
    module: {
        myEntities: new MyEntitiesApi(myEntitiesValidator, myEntitiesMapper),
    },
};
```

### 6) React Hooks (API Hooks)

- Add a hook under `api/hooks/<module>` that consumes the context and exposes:
    - `queries`: async functions that call service methods and return unwrapped results
    - `mutations`: async functions for write operations with error notifications

Example:

```ts
import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useMyEntitiesApi() {
        const { api } = use(MyApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findMany: async (request, signal) => {
                    const result = await api.module.myEntities.findMany({ data: request }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                createOne: async request => {
                    const result = await api.module.myEntities.createOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useMyEntitiesApi = createHook();
```

### 7) Query Keys

- Add query key constants in `module/data/constants/...query-keys.ts`.
- Never use string literals directly in `useQuery`/`useMutation`.

Example:

```ts
export const QK = {
    "my-entities.find-many-paginated": "my-entities.find-many-paginated",
} as const;
```

### 8) Queries (TanStack Query)

- Create a queries file under `module/data/queries/...` that:
    - Uses the API hook’s `queries`
    - Sets `queryKey` from QK and `placeholderData: keepPreviousData`

Example:

```ts
import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useMyEntitiesApi } from "~/<module>/api";
import { QK } from "~/<module>/data/constants";

type FindManyReq = MyEntities_FindMany_Req["data"]; // from contracts
type FindManyRes = MyEntities_FindMany_Res;

function useFindManyPaginated(
    request: FindManyReq,
    options: Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useMyEntitiesApi();
    return useQuery({
        queryKey: [QK["my-entities.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const MyEntitiesQueries = Object.freeze({ useFindManyPaginated });
```

### 9) Commands (TanStack Mutations)

- Create commands under `module/data/commands/...` that:
    - Use the API hook’s `mutations`
    - On success, invalidate (or update) relevant queries

Example:

```ts
import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { QK } from "~/<module>/data/constants";

function useCreateOne(options: Omit<UseMutationOptions<CreateRes, Error, CreateReq>, "mutationFn"> = {}) {
    const { mutations } = useMyEntitiesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, request, context) => {
            void queryClient.invalidateQueries({ queryKey: [QK["my-entities.find-many-paginated"]] });
            options.onSuccess?.(response, request, context);
        },
        ...options,
    });
}

export const MyEntitiesCommands = Object.freeze({ useCreateOne });
```

### 10) Barrel Exports

- Ensure all new files are exported in their local `index.ts` and aggregated indexes for services, hooks, queries, and commands to provide a clean public surface.

---

## Rules and Best Practices

- **Validation**: Always use shared schemas like `PagingMetaApiSchema`. Avoid custom pagination schemas.
- **Dates**: Use `z.coerce.date()` for incoming dates; optional/nullable per requirements.
- **Result Pattern**: Services must return `Result` (`Ok`/`Err`); hooks unwrap with `match` and rethrow/notify.
- **Query Keys**: Centralize keys in `data/constants`. Never hardcode literals in hooks/queries/commands.
- **No Data Conversion in Data Layer**: Do not map/transform inside queries/commands; handle conversion in UI/forms, or in **API service mappers** (`*.api.mapper.ts`) invoked from the service. Keep TanStack layers thin.
- **Module Isolation**: Keep module code within the module. Move shared code to `src/application/shared` or compose existing primitives from `src/components/ui`.

---

## Minimal File Path Template

- `src/application/modules/<module>/api/services/<area>/<entity>/<entity>.api.contracts.ts`
- `src/application/modules/<module>/api/services/<area>/<entity>/<entity>.api.validator.ts`
- `src/application/modules/<module>/api/services/<area>/<entity>/<entity>.api.mapper.ts` (optional; domain ↔ wire)
- `src/application/modules/<module>/api/services/<area>/<entity>/<entity>.api.ts`
- `src/application/modules/<module>/api/services/<area>/<entity>/index.ts`
- `src/application/modules/<module>/api/api-context/<module>.api.context.ts`
- `src/application/modules/<module>/api/hooks/<area>/use-<entity>.api.ts`
- `src/application/modules/<module>/data/constants/<module>.query-keys.ts`
- `src/application/modules/<module>/data/queries/<area>/<entity>.queries.ts`
- `src/application/modules/<module>/data/commands/<area>/<entity>.commands.ts`
- Index barrels under each folder as appropriate

---

## QA Checklist

- Queries resolve with correct types; validators enforce shapes.
- Lint/typecheck pass; UI uses the existing local primitives and shared components.
- Query keys added and used consistently.
- Commands invalidate or update caches appropriately.
- Public barrel exports are complete and organized.

---

## WebSocket Services

Use this section when adding a real-time WebSocket stream alongside (or instead of) a regular REST surface.

### Pattern overview

```
REST service (getToken) → WsApi.streamXxx → Result<WebSocketSubscription, Error>
                                                  ↑
                               BaseWebSocketApi (this.client.buildUrl / connect)
```

WebSocket connections cannot carry HTTP headers in the browser. The standard pattern is:

1. Call a REST endpoint (via an existing `BaseApi` service) to obtain a short-lived `?token=` query parameter.
2. Build the URL and open the socket using `this.client.buildUrl` + `this.client.connect`.
3. Return `Result<WebSocketSubscription, Error>` — same `Result`/`match` discipline as regular services.
4. Wrap `this.client.connect` in a `try/catch` and convert with `toWebSocketError` from `@infrastructure/websocket` (synchronous throw guard for malformed URLs).

### Step-by-step

**1) WS service** — extend `BaseWebSocketApi` in a `*.ws-api.ts` file next to the related REST service:

```ts
import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result, match } from "oxide.ts";

export class MyEntityLogsWsApi extends BaseWebSocketApi {
    public constructor(private readonly myEntityApi: MyEntityApi) {
        super();
    }

    async streamLogs(
        request: MyEntity_GetLogsToken_Req,
        handlers: WebSocketHandlers,
        signal?: AbortSignal,
    ): Promise<Result<WebSocketSubscription, Error>> {
        const tokenResult = await this.myEntityApi.getLogsToken(request, signal);

        return match(tokenResult, {
            Ok: tokenResponse => {
                try {
                    const url = this.client.buildUrl(`my-entity/${encodeURIComponent(request.data.id)}/logs`, {
                        token: tokenResponse.data.token,
                        follow: true,
                    });
                    return Ok(this.client.connect(url, handlers, { signal, closeOnError: true }));
                } catch (error) {
                    return Err(toWebSocketError(error, "Failed to stream logs."));
                }
            },
            Err: error => Err(error),
        });
    }
}
```

**2) Barrel export** — add the WS service to the area `index.ts` alongside existing REST services.

**3) Context wiring** — register as a nested `logs.$` node under the relevant branch:

```ts
myEntity: {
    $: myEntityApi,
    logs: {
        $: new MyEntityLogsWsApi(myEntityApi),
    },
},
```

**4) Hook** — expose `streams.subscribe` (analogous to REST `queries`/`mutations`):

```ts
function createHook() {
    return function useMyEntityLogsWsApi() {
        const { api } = use(MyApiContext);

        const streams = useMemo(
            () => ({
                subscribe: async (
                    data: MyEntity_GetLogsToken_Req["data"],
                    handlers: WebSocketHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await api.myEntity.logs.$.streamLogs({ data }, handlers, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        return { streams };
    };
}

export const useMyEntityLogsWsApi = createHook();
```

**5) Consumer component** — use an `AbortController` + `useEffect` cleanup to ensure the socket is closed when the component unmounts or parameters change:

```ts
useEffect(() => {
    let isDisposed = false;
    let subscription: WebSocketSubscription | null = null;
    const abortController = new AbortController();

    void streams
        .subscribe(
            { id },
            {
                onMessage: message => {
                    /* handle message */
                },
                onReadyStateChange: readyState => {
                    if (!isDisposed) setReadyState(readyState);
                },
            },
            abortController.signal,
        )
        .then(sub => {
            if (isDisposed) {
                sub.close();
                return;
            }
            subscription = sub;
            setReadyState(sub.getReadyState());
        })
        .catch((error: unknown) => {
            if (!isDisposed) console.error("WS connect failed", error);
        });

    return () => {
        isDisposed = true;
        abortController.abort();
        subscription?.close();
    };
}, [id, streams]);
```

### Naming conventions

| Artefact                | Suffix / pattern                                                 |
| ----------------------- | ---------------------------------------------------------------- |
| WS service              | `<entity>.ws-api.ts`                                             |
| WS hook                 | `use-<entity>.ws-api.ts`                                         |
| Handler types re-export | `type MyEntityWsHandlers = WebSocketHandlers`                    |
| Request type alias      | `MyEntity_StreamXxx_Req = <RestTokenReq>` (reuse REST contracts) |
| Response type alias     | `MyEntity_StreamXxx_Res = WebSocketSubscription`                 |

### Rules

- WS services must return `Result<WebSocketSubscription, Error>` — never throw from a service method.
- Always call `toWebSocketError` in the `catch` block around `this.client.connect`.
- The hook exposes a `streams` object (not `queries`/`mutations`) to signal real-time semantics.
- Token acquisition flows through an existing `BaseApi` REST service so the axios auth interceptor (token refresh) still runs.
- No TanStack Query wrapping — WS streams are imperative connections, not cached queries.
