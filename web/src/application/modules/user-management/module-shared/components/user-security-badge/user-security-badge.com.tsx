import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";

import { ESecuritySettings } from "@application/shared/enums";

function View({ securityOption }: Props) {
    const securityMap: Record<ESecuritySettings, string> = {
        [ESecuritySettings.PasswordOnly]: "Password Only",
        [ESecuritySettings.Password2FA]: "Password 2FA",
        [ESecuritySettings.EnforceSSO]: "Enforce SSO",
    };

    const securityToneMap: Partial<Record<string, BadgeTone>> = {
        [ESecuritySettings.PasswordOnly]: "orange",
        [ESecuritySettings.Password2FA]: "green",
        [ESecuritySettings.EnforceSSO]: "blue",
    };

    return (
        <Badge tone={securityToneMap[securityOption] ?? "neutral"}>
            {securityMap[securityOption] || securityOption}
        </Badge>
    );
}

interface Props {
    securityOption: ESecuritySettings;
}

export const UserSecurityBadge = memo(View);
