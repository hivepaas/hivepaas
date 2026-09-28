import { formBox } from "@lib/styles";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { BackupSnapshotRestoreRoute } from "~/settings/module-shared/components";

/** A snapshot of the project's view, or of an env's when one is selected, restored. */
export function ProjectBackupSnapshotRestoreRoute() {
    const { id: projectId, snapshotId = "" } = useParams<{ id: string; snapshotId: string }>();

    invariant(projectId, "projectId must be defined");
    const scopedEnv = getProjectEnvFilterParam(useSelectedProjectEnv(projectId));

    return (
        <div className={formBox}>
            <BackupSnapshotRestoreRoute
                scope={{ type: "project", projectId, env: scopedEnv }}
                snapshotId={snapshotId}
            />
        </div>
    );
}
