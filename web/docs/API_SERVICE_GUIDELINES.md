## API Service Guidelines

This guideline documents the standard process to build a complete API surface for any entity in COMPASS-V2. It covers service and validator implementation, wiring into the API context, React hooks, data queries, and commands. Follow alongside `docs/DEVELOPMENT_GUIDELINES.md` for cross-cutting rules (imports, Zod, Result pattern, query keys, etc.).

### Scope and Outcomes

- **Service**: Type-safe HTTP client methods using `BaseApi` + `Result` pattern.
- **Validator**: Zod schemas using existing meta schemas (e.g., `PagingMetaApiSchema`).
- **Context**: Register the service in the module’s API context factory.
- **Hooks**: React hooks providing `queries` and `mutations` that unwrap `Result` and surface errors with notifications.
- **Queries**: TanStack Query consumers for read operations with proper query keys and placeholder data.
- **Commands**: TanStack Mutation consumers for write operations with cache invalidation.
- **Exports**: Update all related `index.ts` barrels.

---

## Step-by-step Checklist

### 1) Define Contracts (types)

- Create a contracts file under the module service folder. Include:
    - Entity shape (domain-aligned fields)
    - Request/response types based on `ApiRequestBase`/`ApiResponseBase`/`ApiResponsePaginated`
    - Include `WorkspaceMeta` in requests (via `ApiRequestBase<..., WorkspaceMeta>`) and support `pagination`/`sorting`/`search` as needed

Example (concise):

```ts
import type { WorkspaceMeta } from "@application/shared/api/types";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

import type { PaginationState, SortingState } from "@infrastructure/data";

export interface MyEntity {
    id: string;
    name: string;
}

export type MyEntities_FindMany_Req = ApiRequestBase<
    { pagination?: PaginationState; sorting?: SortingState; search?: string },
    WorkspaceMeta
>;
export type MyEntities_FindMany_Res = ApiResponsePaginated<MyEntity>;

export type MyEntities_CreateOne_Req = ApiRequestBase<{ payload: { name: string } }, WorkspaceMeta>;
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

### 3) Implement Service (BaseApi + Result)

- Create a service class extending `BaseApi` that:
    - Uses `this.client.v1` for HTTP
    - Sets `WORKSPACE_HEADER` for requests
    - Builds query params via `this.queryBuilder.getInstance()`
    - Returns `Result` (`Ok`/`Err`) and transforms via validator

Example methods:

```ts
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { WORKSPACE_HEADER } from "@config";

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
                    headers: { [WORKSPACE_HEADER]: req.meta.workspaceId },
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
- Instantiate the validator and add your service into the returned `api` object under the appropriate subtree.

Example (excerpt):

```ts
import { MyEntitiesApi, MyEntitiesApiValidator } from "~/module/api/services";

const myEntitiesValidator = new MyEntitiesApiValidator();

return {
  module: {
    myEntities: new MyEntitiesApi(myEntitiesValidator),
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
import invariant from "tiny-invariant";

import { useWorkspaceId } from "@application/shared/hooks/workspaces";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useMyEntitiesApi() {
        const { workspaceId } = useWorkspaceId();
        invariant(workspaceId, "workspaceId must be defined");

        const { api } = use(MyApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findMany: async (request, signal) => {
                    const result = await api.module.myEntities.findMany(
                        { data: request, meta: { workspaceId } },
                        signal,
                    );
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api, workspaceId],
        );

        const mutations = useMemo(
            () => ({
                createOne: async request => {
                    const result = await api.module.myEntities.createOne({ data: request, meta: { workspaceId } });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError, workspaceId],
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

import { useMyEntitiesApi } from "~/module/api";

import { QK } from "~/module/data/constants";

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

import { QK } from "~/module/data/constants";

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
- **Workspace Meta**: Always pass `workspaceId` via `WORKSPACE_HEADER` and `ApiRequestBase` meta.
- **Query Keys**: Centralize keys in `data/constants`. Never hardcode literals in hooks/queries/commands.
- **No Data Conversion in Data Layer**: Do not map/transform inside queries/commands; handle conversion in UI/forms or dedicated mappers before calling the data layer.
- **Module Isolation**: Keep module code within the module. Move shared code to `src/application/shared` or `src/ui-kit`.

---

## Minimal File Path Template

- `src/application/modules/<module>/api/services/<area>/<entity>/<entity>.api.contracts.ts`
- `src/application/modules/<module>/api/services/<area>/<entity>/<entity>.api.validator.ts`
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
- All services set `WORKSPACE_HEADER` and use QueryBuilder.
- Lint/typecheck pass; no direct `antd` imports (use `@ui-kit`).
- Query keys added and used consistently.
- Commands invalidate or update caches appropriately.
- Public barrel exports are complete and organized.
