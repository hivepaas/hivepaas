import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { AuditLogScope } from "~/operations/domain";
import { AuditLogsList } from "~/operations/routes/audit-logs";
import { ProjectsQueries } from "~/projects/data";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

export function ProjectAuditLogsRoute() {
    const { id: projectId = "" } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? [];

    const scope: AuditLogScope = scopedEnv
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
            <AuditLogsList
                key={`${projectId}-${scopedEnv ?? "all"}`}
                scope={scope}
            />
        </div>
    );
}
