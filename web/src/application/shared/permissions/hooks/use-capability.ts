import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";

/**
 * Checks whether the currently authenticated user holds the specified capability.
 * Admin users hold wildcard access across all capabilities.
 */
export function useCapability(capabilityId: string): { hasCapability: boolean } {
    const profile = useProfileContext(state => state.profile);

    if (!profile) {
        return { hasCapability: false };
    }

    if (profile.role === EUserRole.Admin) {
        return { hasCapability: true };
    }

    return {
        hasCapability: profile.capabilities.includes(capabilityId),
    };
}
