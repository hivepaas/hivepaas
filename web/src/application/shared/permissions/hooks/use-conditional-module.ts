import { useCallback, useMemo } from "react";

import { useParams } from "react-router";

import { MODULE_IDS } from "@application/shared/constants";

import { useModulePermissionsStore, useProjectPermissionsStore } from "../store";
import type {
    ModuleAction,
    ModuleId,
    ModulePermission,
    UseConditionalModuleParams,
    UseConditionalModuleResult,
} from "../types";
import { DENIED_ACTIONS, hasModuleActionAccess } from "../utils";

export function useConditionalModule<const T extends ModuleId>({
    id,
}: UseConditionalModuleParams<T>): UseConditionalModuleResult<T> {
    const { id: routeProjectId } = useParams<{ id: string }>();
    const modules = useModulePermissionsStore(state => state.modules);
    const projects = useProjectPermissionsStore(state => state.projects);

    const module = useMemo(
        () => modules.find((item): item is ModulePermission<T> => item.moduleId === id) ?? null,
        [id, modules],
    );
    const project = useMemo(
        () =>
            id === MODULE_IDS.Project && routeProjectId
                ? (projects.find(item => item.projectId === routeProjectId) ?? null)
                : null,
        [id, projects, routeProjectId],
    );

    const actions = project?.actions ?? module?.actions ?? DENIED_ACTIONS;

    const hasAccess = useCallback((action: ModuleAction) => hasModuleActionAccess(actions, action), [actions]);

    return {
        module,
        actions,
        canRead: hasAccess("read"),
        canExecute: hasAccess("execute"),
        canWrite: hasAccess("write"),
        canDelete: hasAccess("delete"),
        hasAccess,
    };
}
