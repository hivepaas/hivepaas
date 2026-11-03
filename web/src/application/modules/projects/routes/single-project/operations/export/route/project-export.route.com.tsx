import { useMemo } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { SpecExportScope, SpecImportScope } from "~/operations/domain";
import { SpecExportPanel, SpecImportPanel } from "~/operations/routes/export";
import { ProjectsQueries } from "~/projects/data";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { MODULE_IDS } from "@application/shared/constants";

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
    // Kept stable across renders: the import panel asks for its plan again when
    // its scope changes.
    const importScope = useMemo<SpecImportScope>(
        () =>
            scopedEnv
                ? { type: "project-env", projectID: projectId, projectEnvID: scopedEnv }
                : { type: "project", projectID: projectId },
        [projectId, scopedEnv],
    );

    const scopeLabel = scopedEnv ? `${projectName} (${scopedEnv})` : projectName;

    return (
        <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2">
                <ProjectEnvScopeBadge
                    selectedEnv={selectedEnv}
                    envs={projectEnvs}
                />
            </div>
            <div className="flex flex-col gap-3">
                <h2 className="text-base font-semibold text-foreground">Export</h2>
                <SpecExportPanel
                    key={`${projectId}-${scopedEnv ?? "all"}`}
                    scope={scope}
                    scopeLabel={scopeLabel}
                />
            </div>
            <div className="flex flex-col gap-3">
                <h2 className="text-base font-semibold text-foreground">Import</h2>
                <SpecImportPanel
                    key={`${projectId}-${scopedEnv ?? "all"}`}
                    scope={importScope}
                    scopeLabel={scopeLabel}
                    permissionModuleId={MODULE_IDS.Project}
                />
            </div>
        </div>
    );
}
