import { inlineLink } from "@lib/styles";
import { cn } from "@lib/utils";
import { ExternalLink } from "lucide-react";

/*
 * What Bee leaves a Google, Microsoft or OpenID Connect sign-in provider, said
 * the same way on the list, the form and the status dialog.
 *
 * Bee allows none to be added or enabled. One that is active keeps signing people
 * in - turning it off would lock out those who have no password - and stays
 * editable, so an expiring client secret can still be replaced. Once disabled, it
 * stays so until a paid plan is in force. The server enforces all of it; these
 * say so before anyone runs into its refusal.
 */

/** Said of one that is active on Bee. */
export const OAUTH_KEPT_ON_BEE_NOTE =
    "It keeps signing people in on Bee. Once disabled, it can't be enabled again without a paid plan.";

/** Said of one that is not active on Bee, beside what would enable it. */
export const OAUTH_ENABLE_NEEDS_PAID_PLAN = "A paid plan is needed to enable it.";

/** Said of one that is not active on Bee, anywhere else. */
export const OAUTH_UNAVAILABLE_ON_BEE_NOTE = "It is unavailable on Bee: a paid plan is needed to enable it.";

/**
 * A link to the license portal's pricing page, which carries this installation's
 * ID. Nothing when the portal's address is not known.
 */
export function PaidPlanUpgradeLink({ pricingUrl }: Props) {
    if (!pricingUrl) {
        return null;
    }

    return (
        <a
            href={pricingUrl}
            target="_blank"
            rel="noopener noreferrer"
            className={cn(inlineLink, "inline-flex items-center gap-1")}
        >
            Upgrade
            <ExternalLink className="size-3 shrink-0" />
        </a>
    );
}

interface Props {
    /** The license's portal.pricingUrl. */
    pricingUrl?: string;
}
