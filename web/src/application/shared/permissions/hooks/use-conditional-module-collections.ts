import { useMemo } from "react";

import { useModulePermissionsStore } from "../store";
import type { ModuleId, ModulePermission } from "../types";

interface UseConditionalModuleCollectionsResult {
    list: readonly ModulePermission[];
    map: ReadonlyMap<ModuleId, ModulePermission>;
}

export function useConditionalModuleCollections(): UseConditionalModuleCollectionsResult {
    const modules = useModulePermissionsStore(state => state.modules);

    return useMemo(() => {
        const list = Object.freeze([...modules]);
        const map: ReadonlyMap<ModuleId, ModulePermission> = new Map(modules.map(module => [module.moduleId, module]));

        return Object.freeze({
            list,
            map,
        });
    }, [modules]);
}
