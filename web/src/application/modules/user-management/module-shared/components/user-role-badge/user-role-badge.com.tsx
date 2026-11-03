import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";

import { EUserRole } from "@application/shared/enums";

function View({ role }: Props) {
    const roleMap: Record<EUserRole, string> = {
        [EUserRole.Admin]: "Admin",
        [EUserRole.Member]: "Member",
    };

    const roleToneMap: Partial<Record<string, BadgeTone>> = {
        [EUserRole.Admin]: "amber",
        [EUserRole.Member]: "purple",
    };

    return <Badge tone={roleToneMap[role] ?? "neutral"}>{roleMap[role] || role}</Badge>;
}

interface Props {
    role: EUserRole;
}

export const UserRoleBadge = memo(View);
