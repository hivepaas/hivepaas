import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { ProjectCommandTemplateFormRoute } from "../form-route";

export function ProjectCommandTemplateCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <ProjectCommandTemplateFormRoute
            mode="create"
            projectId={projectId}
            env={scopedEnv}
        />
    );
}
