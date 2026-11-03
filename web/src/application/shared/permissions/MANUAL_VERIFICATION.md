# Module Permission Manual Verification

- Member without a module permission cannot see that module in the sidebar and receives the no-access page when opening the route directly.
- Missing module or missing action resolves to `false`; `canRead`, `canWrite`, and `canDelete` are `false`.
- Member with `read: true` for a module resolves `canRead` to `true` and can render `ConditionalModule` for the read action.
- Member with `write: false` for a module renders `fallback` for `<ConditionalModule action="write" />`.
- Admin receives full `read`, `write`, and `delete` access for every `MODULES` entry, regardless of returned module access rows.
- Logout or session invalidation clears profile and module permissions, returning the app to deny-by-default behavior.
