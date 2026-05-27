# LocalPaaS Dashboard Development Guidelines

## Project Root

Work from `localpaas-dashboard/` for frontend tasks.

Common scripts:

- `npm run dev`
- `npm run lint:ci`
- `npm run build`
- `npm run format:src`
- `npm run codegen`
- `npm run agents:check`

## Agent Quick Start

For Codex/Cursor frontend work:

- Start from `localpaas-dashboard/`.
- Read this file first, then load only the matching `.cursor/rules/*.mdc`.
- Use `$fe-dev` for frontend implementation/review.
- Use `$convention-enforcer` before and after frontend edits.
- Use `$api-layer-review` for API services, validators, hooks, query keys, queries, and commands.
- Run `npm run agents:check` after changing skills, Cursor rules, or agent-facing docs.

## Rule Map

- `frontend-architecture.mdc`: module ownership, folder shape, imports, TypeScript/React, styling.
- `api-data-layer.mdc`: contracts, validators, services, API hooks, query keys, queries, commands, generated API.
- `routing-dialogs-ui.mdc`: route wiring, navigation wrappers, dialogs, forms, tables, UI composition.
- `quality-checks.mdc`: validation commands, dependency boundaries, generated files, formatting scope.
- Root `.cursor/rules/localpaas-dashboard-fe.mdc`: parent-workspace pointer to this app and these rules.

## Architecture

- `src/application/modules/<module>` contains feature modules such as projects, settings, sources, cluster, system-settings, and user-management.
- `src/application/authentication` owns authentication routes, context, API, and data.
- `src/application/shared` is for cross-module application components, layouts, pages, dialogs, hooks, data, constants, and entities.
- `src/infrastructure` is for API clients, validation, exceptions, generic data types, device/services, and other framework-level utilities.
- `src/components/ui` contains shadcn/Radix-style primitives. Prefer composition before changing a primitive for one screen.

## Module Shape

Use the existing module folders when they apply:

- `api/services`: `*.api.contracts.ts`, `*.api.validator.ts`, `*.api.ts`, and local `index.ts` exports.
- `api/api-context`: module API context and service construction.
- `api/hooks`: hooks that expose `queries` and `mutations`.
- `data/constants`: module query keys.
- `data/queries`: TanStack Query hooks.
- `data/commands`: TanStack Mutation hooks.
- `domain`: domain entities and enums.
- `routes`: route components, forms, schemas, and building blocks.
- `module-shared`: components reused inside the same module.
- `dialogs` and `dialogs-container`: module dialogs and their wiring.
- `layouts`: module layout components.

## Naming

- UI components: `*.com.tsx`
- SCSS modules: `*.module.scss`
- Table definitions: `*.defs.tsx`
- Shared props/types: `*.types.ts`
- API contracts: `*.api.contracts.ts`
- API validators: `*.api.validator.ts`
- API services: `*.api.ts`
- Queries: `*.queries.ts`
- Commands: `*.commands.ts`

Export public files through nearby `index.ts` barrels and module aggregators.

## Imports

- Use `~/<module>/...` for module-owned code under `src/application/modules`.
- Use `@application/shared/...` for cross-module application code.
- Use `@infrastructure/...` for infrastructure code.
- Use `@assets/...` for assets.
- Use `@/components/ui` or existing local UI imports for shared primitives.
- Prefer type-only imports for types.
- Let Prettier sort import groups.

## UI, Forms, Tables, And Dialogs

- Build dashboard screens as dense, scannable operational interfaces.
- Use existing shared/module components before creating new ones.
- Use React Hook Form with Zod for non-trivial forms.
- Use `FieldError` and existing form layout helpers where nearby code does.
- Use `DataTable` with `useTableState` for list tables.
- Keep table columns in `*.defs.tsx` for non-trivial tables.
- Follow `docs/DIALOGS.md` for dialog state, hooks, UI, and container wiring.

## Routing

- Add route constants in `src/application/shared/constants/route.constants.ts`.
- Wire routes through the owning module router.
- Use module `*.module.ts` exports for lazy route loading.
- Use `AppLink`, `AppNavLink`, `AppNavigate`, and `useAppNavigate` wrappers to preserve previous-path state.

## API And Data

Follow `docs/API_SERVICE_GUIDELINES.md` for complete API surfaces.

Key boundaries:

- Components call queries and commands, not API services directly.
- API hooks unwrap `Result` values and surface user-facing errors.
- Services own HTTP, validation, and wire/domain mapping.
- Validators parse and normalize API responses.
- Commands invalidate affected query keys.

## Validation

Use the smallest check that covers the risk:

- `npm run agents:check` after skills/rules/docs setup changes.
- `npm run lint:ci` for TypeScript and ESLint.
- `npm run build` after route/module/API context/build config changes.
- `npm run dev` plus browser verification for visible UI changes.

Report skipped checks with the exact reason.
