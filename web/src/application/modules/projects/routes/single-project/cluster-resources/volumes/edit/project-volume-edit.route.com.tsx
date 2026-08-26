import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { EditVolumeFormRoute } from "~/cluster/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

export function ProjectVolumeEditRoute() {
    const { id: projectId, volumeId } = useParams<{ id: string; volumeId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(volumeId, "volumeId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <EditVolumeFormRoute
            scope={{ type: "project", projectId, env: scopedEnv }}
            volumeId={volumeId}
        />
    );
}
