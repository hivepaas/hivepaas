import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { ProjectCommandPipeFormRoute } from "../form-route";

export function ProjectCommandPipeCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <ProjectCommandPipeFormRoute
            mode="create"
            projectId={projectId}
            env={scopedEnv}
        />
    );
}
