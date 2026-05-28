import { create } from "zustand";

import type { ModuleId, ModulePermission } from "../types";
import { createFullModulePermissions } from "../utils";

interface ModulePermissionsState {
    modules: readonly ModulePermission[];
    setModules: (modules: readonly ModulePermission[]) => void;
    setFullModules: (moduleIds?: readonly ModuleId[]) => void;
    clearModules: () => void;
}

export const useModulePermissionsStore = create<ModulePermissionsState>()(set => ({
    modules: [],

    setModules: modules => {
        set({
            modules: [...modules],
        });
    },

    setFullModules: moduleIds => {
        set({
            modules: createFullModulePermissions(moduleIds),
        });
    },

    clearModules: () => {
        set({
            modules: [],
        });
    },
}));
