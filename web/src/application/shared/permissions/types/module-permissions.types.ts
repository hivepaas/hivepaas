export type ModuleAction = "read" | "execute" | "write" | "delete";

export type ModuleId = string;

export interface ModulePermission<T extends ModuleId = ModuleId> {
    moduleId: T;
    actions: Readonly<Record<ModuleAction, boolean>>;
}

export interface UseConditionalModuleParams<T extends ModuleId> {
    id: T;
}

export interface UseConditionalModuleResult<T extends ModuleId> {
    module: ModulePermission<T> | null;
    actions: ModulePermission["actions"];
    canRead: boolean;
    canExecute: boolean;
    canWrite: boolean;
    canDelete: boolean;
    hasAccess: (action: ModuleAction) => boolean;
}
