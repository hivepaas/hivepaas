import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { CreateVolumeFormRoute } from "~/cluster/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

export function ProjectVolumeCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return <CreateVolumeFormRoute scope={{ type: "project", projectId, env: scopedEnv }} />;
}
