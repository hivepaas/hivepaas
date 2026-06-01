import type { ModuleAction } from "../../types";

const MODULE_PERMISSION_DENIED_MESSAGES: Readonly<Record<ModuleAction, string>> = Object.freeze({
    read: "You do not have permission to view this item.",
    execute: "You do not have permission to run this action.",
    write: "You only have view access. Create or edit actions are disabled.",
    delete: "You do not have permission to delete this item.",
});

export function getModulePermissionDeniedMessage(action: ModuleAction): string {
    return MODULE_PERMISSION_DENIED_MESSAGES[action];
}
