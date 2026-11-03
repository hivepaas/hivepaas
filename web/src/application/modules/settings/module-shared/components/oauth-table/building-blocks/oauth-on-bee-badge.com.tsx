import { memo } from "react";

import { Badge } from "@components/ui/badge";
import { OAUTH_ENABLE_NEEDS_PAID_PLAN } from "~/settings/module-shared/components/oauth-form";

/**
 * What Bee leaves a Google, Microsoft or OpenID Connect provider: one that is
 * active keeps signing people in, and none other can be enabled.
 */
function View({ isActive }: Props) {
    if (isActive) {
        return (
            <Badge
                tone="amber"
                title="It keeps signing people in. No other Google, Microsoft or OpenID Connect provider can be added or enabled on Bee."
            >
                Kept on Bee
            </Badge>
        );
    }

    return (
        <Badge
            tone="neutral"
            title={OAUTH_ENABLE_NEEDS_PAID_PLAN}
        >
            Unavailable on Bee
        </Badge>
    );
}

interface Props {
    isActive: boolean;
}

export const OAuthOnBeeBadge = memo(View);
