import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";

import { ESecuritySettings } from "@application/shared/enums";

function View({ securityOption }: Props) {
    const securityMap: Record<ESecuritySettings, string> = {
        [ESecuritySettings.PasswordOnly]: "Password Only",
        [ESecuritySettings.Password2FA]: "Password + 2FA",
        [ESecuritySettings.EnforceSSO]: "Enforce SSO",
    };

    const securityColorMap: Record<ESecuritySettings, string> = {
        [ESecuritySettings.PasswordOnly]: "bg-orange-400 text-white",
        [ESecuritySettings.Password2FA]: "bg-green-500 text-white",
        [ESecuritySettings.EnforceSSO]: "bg-blue-600 text-white",
    };

    return (
        <Badge className={cn(securityColorMap[securityOption] || "bg-primary text-primary-foreground")}>
            {securityMap[securityOption] || securityOption}
        </Badge>
    );
}

interface Props {
    securityOption: ESecuritySettings;
}

export const UserSecurityBadge = memo(View);
