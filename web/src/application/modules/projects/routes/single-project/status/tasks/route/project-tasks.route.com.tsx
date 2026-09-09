import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import type { SystemTaskScope } from "~/system-status/domain";
import { SystemTasksList } from "~/system-status/routes/tasks";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

export function ProjectTasksRoute() {
    const { id: projectId = "" } = useParams<{ id: string }>();
    const { navigate } = useAppNavigate();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? [];

    const scope: SystemTaskScope = scopedEnv
        ? {
              type: "project-env",
              projectID: projectId,
              projectEnvID: scopedEnv,
          }
        : {
              type: "project",
              projectID: projectId,
          };

    return (
        <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2">
                <ProjectEnvScopeBadge
                    selectedEnv={selectedEnv}
                    envs={projectEnvs}
                />
            </div>
            <SystemTasksList
                key={`${projectId}-${scopedEnv ?? "all"}`}
                scope={scope}
                onSelectTask={task => {
                    navigate.modules(ROUTE.projects.single.status.tasks.details.$route(projectId, task.id));
                }}
            />
        </div>
    );
}
