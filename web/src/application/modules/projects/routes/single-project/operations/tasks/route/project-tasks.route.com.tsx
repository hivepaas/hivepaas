import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { SystemTaskScope } from "~/operations/domain";
import { SystemTasksList } from "~/operations/routes/tasks";
import { ProjectsQueries } from "~/projects/data";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

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
                    navigate.modules(ROUTE.projects.single.operations.tasks.details.$route(projectId, task.id));
                }}
            />
        </div>
    );
}
