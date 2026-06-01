import { useCallback, useMemo } from "react";

import { MODULE_IDS } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";

import { useModulePermissionsStore, useProjectPermissionsStore } from "../store";
import type {
    ModuleAction,
    ModulePermission,
    ProjectPermission,
    UseConditionalProjectParams,
    UseConditionalProjectResult,
} from "../types";
import { DENIED_ACTIONS, FULL_ACTIONS, hasModuleActionAccess } from "../utils";

export function useConditionalProject<const T extends string>({
    projectId,
}: UseConditionalProjectParams<T>): UseConditionalProjectResult<T> {
    const profile = useProfileContext(state => state.profile);
    const projects = useProjectPermissionsStore(state => state.projects);
    const modules = useModulePermissionsStore(state => state.modules);

    const project = useMemo(
        () => projects.find((item): item is ProjectPermission<T> => item.projectId === projectId) ?? null,
        [projectId, projects],
    );

    const module = useMemo(
        () =>
            modules.find(
                (item): item is ModulePermission<typeof MODULE_IDS.Project> => item.moduleId === MODULE_IDS.Project,
            ) ?? null,
        [modules],
    );

    const isAdmin = profile?.role === EUserRole.Admin;
    const actions = isAdmin ? FULL_ACTIONS : (project?.actions ?? module?.actions ?? DENIED_ACTIONS);
    const source = isAdmin ? "admin" : project ? "project" : module ? "module" : "none";

    const hasAccess = useCallback((action: ModuleAction) => hasModuleActionAccess(actions, action), [actions]);

    return {
        project,
        actions,
        canRead: hasAccess("read"),
        canWrite: hasAccess("write"),
        canDelete: hasAccess("delete"),
        hasAccess,
        source,
    };
}
