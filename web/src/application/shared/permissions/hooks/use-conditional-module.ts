import { useCallback, useMemo } from "react";

import { useModulePermissionsStore } from "../store";
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
    const modules = useModulePermissionsStore(state => state.modules);

    const module = useMemo(
        () => modules.find((item): item is ModulePermission<T> => item.moduleId === id) ?? null,
        [id, modules],
    );

    const actions = module?.actions ?? DENIED_ACTIONS;

    const hasAccess = useCallback((action: ModuleAction) => hasModuleActionAccess(actions, action), [actions]);

    return {
        module,
        actions,
        canRead: hasAccess("read"),
        canWrite: hasAccess("write"),
        canDelete: hasAccess("delete"),
        hasAccess,
    };
}
