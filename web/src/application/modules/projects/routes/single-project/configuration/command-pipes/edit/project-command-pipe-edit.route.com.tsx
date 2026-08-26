import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { ProjectCommandPipeFormRoute } from "../form-route";

export function ProjectCommandPipeEditRoute() {
    const { commandPipeId, id: projectId } = useParams<{ commandPipeId: string; id: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(commandPipeId, "commandPipeId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <ProjectCommandPipeFormRoute
            mode="edit"
            projectId={projectId}
            env={scopedEnv}
            commandPipeId={commandPipeId}
        />
    );
}
