import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { ProjectCommandTemplateFormRoute } from "../form-route";

export function ProjectCommandTemplateEditRoute() {
    const { commandTemplateId, id: projectId } = useParams<{ commandTemplateId: string; id: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(commandTemplateId, "commandTemplateId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <ProjectCommandTemplateFormRoute
            mode="edit"
            projectId={projectId}
            env={scopedEnv}
            commandTemplateId={commandTemplateId}
        />
    );
}
