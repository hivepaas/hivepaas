import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { EnvScheduledJobsQueries } from "~/projects/data/queries";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { EnvJobSequenceFormRoute } from "../form-route";
import { ProjectScheduledJobsRoute } from "../route";

export function EnvJobSequenceEditRoute() {
    const { id: projectId, scheduledJobId } = useParams<{ id: string; scheduledJobId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(scheduledJobId, "scheduledJobId must be defined");
    const scopedEnv = getProjectEnvFilterParam(useSelectedProjectEnv(projectId));

    const { data } = EnvScheduledJobsQueries.useFindOneById(
        { projectID: projectId, env: scopedEnv ?? "", scheduledJobID: scheduledJobId },
        { enabled: Boolean(scopedEnv) },
    );

    if (!scopedEnv) {
        return <ProjectScheduledJobsRoute />;
    }

    return (
        <EnvJobSequenceFormRoute
            mode="edit"
            projectId={projectId}
            env={scopedEnv}
            scheduledJob={data?.data}
        />
    );
}
