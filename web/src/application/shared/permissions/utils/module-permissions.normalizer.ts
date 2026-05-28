import { MODULES } from "@application/shared/constants";

import type { ModuleAction, ModuleId, ModulePermission } from "../types";

export const MODULE_ACTIONS = ["read", "write", "delete"] as const satisfies readonly ModuleAction[];

export const DENIED_ACTIONS: ModulePermission["actions"] = Object.freeze({
    read: false,
    write: false,
    delete: false,
});

export const FULL_ACTIONS: ModulePermission["actions"] = Object.freeze({
    read: true,
    write: true,
    delete: true,
});

const DEFAULT_MODULE_IDS = MODULES.map(module => module.id);

export function createFullModulePermissions<T extends ModuleId = ModuleId>(
    moduleIds: readonly T[] = DEFAULT_MODULE_IDS as unknown as readonly T[],
): ModulePermission<T>[] {
    return moduleIds.map(moduleId => ({
        moduleId,
        actions: FULL_ACTIONS,
    }));
}

export function hasModuleActionAccess(actions: ModulePermission["actions"], action: ModuleAction): boolean {
    return actions[action];
}
