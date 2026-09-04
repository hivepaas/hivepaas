import { useCallback } from "react";

import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";

export type AccessAction = "read" | "execute" | "write" | "delete";

/**
 * Reports which permissions the signed-in account is allowed to hand out.
 *
 * A user must not grant - or take away - more than they hold themselves; the API
 * enforces that. This mirrors the rule in the UI so the boxes out of reach are
 * visibly disabled instead of failing on save. Admins are unrestricted.
 *
 * The account's own per-env access comes from the profile, where it is already
 * the effective access (an env with no grant of its own inherits the project
 * level), so no extra resolution is needed here.
 */
export interface GrantableAccess {
    isAdmin: boolean;
    canGrantEnv: (envId: string, action: AccessAction) => boolean;
    canGrantModule: (moduleId: string, action: AccessAction) => boolean;
}

export function useGrantableAccess(): GrantableAccess {
    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const projectAccesses = profile?.projectAccesses;
    const moduleAccesses = profile?.moduleAccesses;

    const canGrantEnv = useCallback(
        (envId: string, action: AccessAction) => {
            if (isAdmin) {
                return true;
            }
            for (const projectAccess of projectAccesses ?? []) {
                const envAccess = projectAccess.envAccesses.find(env => env.id === envId);
                if (envAccess) {
                    return envAccess.access[action];
                }
            }
            return false;
        },
        [isAdmin, projectAccesses],
    );

    const canGrantModule = useCallback(
        (moduleId: string, action: AccessAction) => {
            if (isAdmin) {
                return true;
            }
            const moduleAccess = (moduleAccesses ?? []).find(module => module.id === moduleId);
            return moduleAccess?.access[action] ?? false;
        },
        [isAdmin, moduleAccesses],
    );

    return { isAdmin, canGrantEnv, canGrantModule };
}
