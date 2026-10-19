import { useQueries } from "@tanstack/react-query";
import { useAppManagedApi } from "~/projects/api/hooks/project-apps";
import { QK } from "~/projects/data/constants";
import { ProjectAppsQueries } from "~/projects/data/queries";

/** A managed database of an environment, as a select offers it. */
export interface ManagedDatabaseOption {
    id: string;
    name: string;
}

const LIST_ALL_PAGE = { page: 1, size: 1000 };

/**
 * The environment's databases an engine looks after, whatever the engine, but
 * the app itself: what a deployment can mark a restore point on. The
 * environment's database apps are listed, then each is asked for its engine's
 * settings - with the key the Database tab's query uses, so what was asked is
 * not asked again. One no engine looks after has none.
 */
export function useManagedDatabases(scope: { projectID: string; env: string; appID: string }): {
    databases: ManagedDatabaseOption[];
    isLoading: boolean;
} {
    const { queries } = useAppManagedApi();
    const { data: apps, isLoading: isLoadingApps } = ProjectAppsQueries.useFindManyPaginated({
        projectID: scope.projectID,
        env: scope.env,
        category: ["database"],
        pagination: LIST_ALL_PAGE,
    });
    const candidates = (apps?.data ?? []).filter(app => app.id !== scope.appID);
    const engines = useQueries({
        queries: candidates.map(app => {
            const request = { projectID: scope.projectID, env: scope.env, appID: app.id };
            return {
                queryKey: [QK["projects.apps.managed.$.find-settings"], request],
                queryFn: ({ signal }: { signal: AbortSignal }) => queries.findSettings(request, signal),
                retry: false,
            };
        }),
    });

    const databases = candidates
        .filter((_, index) => engines[index]?.data !== undefined)
        .map(app => ({ id: app.id, name: app.name }));
    return { databases, isLoading: isLoadingApps || engines.some(engine => engine.isLoading) };
}
