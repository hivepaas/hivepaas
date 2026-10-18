import { AppManagedQueries, ProjectAppsQueries } from "~/projects/data/queries";
import type { AppManagedLocks } from "~/projects/domain";

/** What a managed database's engine owns of it: its settings screens show it read-only. */
export interface AppManagedOwnership {
    engine: string;
    locks: AppManagedLocks;
}

/**
 * The engine's locks on an app, null for an app no engine looks after. Only a
 * database can be managed: any other app is not asked. The app is the header's
 * query, already loaded.
 */
export function useAppManagedLocks(scope: {
    projectID: string;
    env: string;
    appID: string;
}): AppManagedOwnership | null {
    // Both are the header's, kept fresh there: a screen reads them as they are.
    const { data: app } = ProjectAppsQueries.useFindOneById({ ...scope, getStats: true }, { refetchOnMount: false });
    const { data } = AppManagedQueries.useFindSettings(scope, {
        enabled: app?.data.category === "database",
        refetchOnMount: false,
    });

    return data ? { engine: data.data.engine, locks: data.data.locks } : null;
}
