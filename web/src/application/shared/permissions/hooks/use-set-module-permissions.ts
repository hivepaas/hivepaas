import { useModulePermissionsStore } from "../store";
import type { ModuleId, ModulePermission } from "../types";

interface UseSetModulePermissionsResult {
    setModulePermissions: (modules: readonly ModulePermission[]) => void;
    setFullModulePermissions: (moduleIds?: readonly ModuleId[]) => void;
    clearModulePermissions: () => void;
}

export function useSetModulePermissions(): UseSetModulePermissionsResult {
    const setModules = useModulePermissionsStore(state => state.setModules);
    const setFullModules = useModulePermissionsStore(state => state.setFullModules);
    const clearModules = useModulePermissionsStore(state => state.clearModules);

    return {
        setModulePermissions: setModules,
        setFullModulePermissions: setFullModules,
        clearModulePermissions: clearModules,
    };
}
