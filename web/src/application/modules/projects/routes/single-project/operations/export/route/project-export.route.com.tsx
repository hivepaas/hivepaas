import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { SpecExportScope } from "~/operations/domain";
import { SpecExportPanel } from "~/operations/routes/export";
import { ProjectsQueries } from "~/projects/data";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

/**
 * Scope follows the environment picker in the top right rather than adding a
 * second one. With no environment selected the whole project is exported; with
 * one selected, just that environment - the same rule the audit log view uses,
 * so the picker means one thing everywhere in a project.
 */
export function ProjectExportRoute() {
    const { id: projectId = "" } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? [];
    const projectName = projectData?.data.name ?? "this project";

    const scope: SpecExportScope = scopedEnv
        ? { type: "project-env", projectID: projectId, projectEnvID: scopedEnv }
        : { type: "project", projectID: projectId };

    const scopeLabel = scopedEnv ? `${projectName} (${scopedEnv})` : projectName;

    return (
        <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2">
                <ProjectEnvScopeBadge
                    selectedEnv={selectedEnv}
                    envs={projectEnvs}
                />
            </div>
            <SpecExportPanel
                key={`${projectId}-${scopedEnv ?? "all"}`}
                scope={scope}
                scopeLabel={scopeLabel}
            />
        </div>
    );
}
