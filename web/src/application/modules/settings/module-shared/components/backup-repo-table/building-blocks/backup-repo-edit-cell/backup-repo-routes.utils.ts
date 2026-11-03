import { ROUTE } from "@application/shared/constants";

import type { BackupRepoTableScope } from "../../backup-repo-table.types";

export function getBackupRepoEditRoute(scope: BackupRepoTableScope, id: string): string {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.backupRepos.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.backupRepos.edit.$route(id);
}

export function getBackupRepoCreateRoute(scope: BackupRepoTableScope): string {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.backupRepos.create.$route(scope.projectId);
    }

    return ROUTE.settings.backupRepos.create.$route;
}
