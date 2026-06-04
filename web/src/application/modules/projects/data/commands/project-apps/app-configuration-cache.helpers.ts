import type { QueryClient } from "@tanstack/react-query";

import { QK } from "../../constants/projects.query-keys";

interface SingleAppCacheScope {
    projectID: string;
    appID: string;
}

export function invalidateSingleAppSummaryQueries(queryClient: QueryClient, scope: SingleAppCacheScope): void {
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.$.find-many-paginated"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.$.find-one-by-id"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.$.find-one-by-id"], { projectID: scope.projectID }],
    });
}

export function invalidateSingleAppConfigurationQueries(queryClient: QueryClient, scope: SingleAppCacheScope): void {
    invalidateSingleAppSummaryQueries(queryClient, scope);

    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.env-vars.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.container-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.deployment-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.service-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.network-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.resource-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.storage-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.http-settings.$.find-one"], scope],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.secrets.$.find-many-paginated"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.secrets.$.find-one-by-id"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.config-files.$.find-many-paginated"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.config-files.$.find-one-by-id"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.health-checks.$.find-many-paginated"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.health-checks.$.find-one-by-id"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-one-by-id"]],
    });
}
