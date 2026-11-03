import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { EnvJobSequenceFormRoute } from "../form-route";
import { ProjectScheduledJobsRoute } from "../route";

export function EnvJobSequenceCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const scopedEnv = getProjectEnvFilterParam(useSelectedProjectEnv(projectId));

    // A sequence belongs to an env: without one picked, the list says so.
    if (!scopedEnv) {
        return <ProjectScheduledJobsRoute />;
    }

    return (
        <EnvJobSequenceFormRoute
            key={scopedEnv}
            mode="create"
            projectId={projectId}
            env={scopedEnv}
        />
    );
}
