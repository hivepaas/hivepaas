import { AlertTriangle } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

/**
 * Tells an admin signing in with a password alone that it is about to stop being
 * enough.
 *
 * Shown to admins only, and only while their own account is on "Password Only".
 * A member is not held to this rule, so warning them would be crying wolf - and a
 * banner that turns out not to apply is one people learn to scroll past, which
 * costs the next one that does apply.
 *
 * The claim is deliberately narrow. The server refuses a change that would leave
 * an admin on a password alone, and grandfathers the accounts already in that
 * state - the bootstrap admin is seeded that way - so nothing is broken today.
 * "Changes to your role or security option are refused" is the whole of what is
 * true, and it is what the banner says.
 *
 * Two links because it takes two pages. Activating TOTP lives on the profile;
 * the security option itself is a field on the user record, which only the user
 * management form writes. Sending somebody to the profile alone would be sending
 * them to do half the job and come back to the same banner.
 */
export function WeakAccountSecurityWarning() {
    const profile = useProfileContext(state => state.profile);

    if (profile === null) {
        return null;
    }
    if (profile.role !== EUserRole.Admin || profile.securityOption !== ESecuritySettings.PasswordOnly) {
        return null;
    }

    const linkClass = "font-medium text-primary underline underline-offset-4 hover:opacity-80";

    return (
        <div
            role="status"
            className="flex items-start gap-2.5 rounded-lg border border-amber-500/30 bg-amber-500/[0.07] px-3.5 py-2.5 dark:border-amber-400/25 dark:bg-amber-400/[0.07]"
        >
            <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
            <p className="text-sm text-foreground">
                <span className="font-medium">Your admin account signs in with a password alone.</span>{" "}
                <span className="text-muted-foreground">
                    An admin can grant themselves anything, so changes to an admin&apos;s role or security option are
                    refused while it stays that way. Activate two-factor authentication on your
                </span>{" "}
                <AppLink.Basic
                    to={ROUTE.currentUser.profile.$route}
                    className={linkClass}
                >
                    profile
                </AppLink.Basic>
                <span className="text-muted-foreground">, then switch the security option on</span>{" "}
                <AppLink.Basic
                    to={ROUTE.userManagement.users.single.$route(profile.id)}
                    className={linkClass}
                >
                    your account
                </AppLink.Basic>
                <span className="text-muted-foreground">.</span>
            </p>
        </div>
    );
}
