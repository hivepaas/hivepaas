import { formBox } from "@lib/styles";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { BackupSnapshotRestoreRoute } from "~/settings/module-shared/components";

/** One of an app's snapshots, restored. */
export function AppBackupSnapshotRestoreRoute() {
    const {
        id: projectId,
        env,
        appId,
        snapshotId = "",
    } = useParams<{ id: string; env: string; appId: string; snapshotId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    return (
        <div className={formBox}>
            <BackupSnapshotRestoreRoute
                scope={{ type: "app", projectId, env, appId }}
                snapshotId={snapshotId}
            />
        </div>
    );
}
