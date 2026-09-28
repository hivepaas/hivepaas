import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { BackupSnapshotTable } from "~/settings/module-shared/components";

/** An app's own snapshots, from every repository it sees. */
export function AppBackupSnapshotsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    return <BackupSnapshotTable scope={{ type: "app", projectId, env, appId }} />;
}
