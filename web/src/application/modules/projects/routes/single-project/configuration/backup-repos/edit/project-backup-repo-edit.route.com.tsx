import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { BackupRepoFormRoute } from "~/settings/module-shared/components/backup-repo-form-route";

export function ProjectBackupRepoEditRoute() {
    const { id: projectId = "", backupRepoId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <BackupRepoFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            backupRepoId={backupRepoId}
        />
    );
}
