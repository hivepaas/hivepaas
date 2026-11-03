import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectProviderSettingsScopeHeader } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { EnvScheduledJobsTable } from "../building-blocks";

/** An env's scheduled jobs: there are none at the project scope, so an env has to be picked. */
export function ProjectScheduledJobsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader
                projectId={projectId}
                hideNote
            />
            {scopedEnv ? (
                <EnvScheduledJobsTable
                    key={scopedEnv}
                    projectId={projectId}
                    env={scopedEnv}
                />
            ) : (
                <div className={cn(dashedBorderBox, "w-fit text-sm leading-normal")}>
                    Scheduled jobs belong to an env and its apps: pick an env in the top right to see them, or to create
                    a job sequence that runs the jobs of its apps in order.
                </div>
            )}
        </div>
    );
}
