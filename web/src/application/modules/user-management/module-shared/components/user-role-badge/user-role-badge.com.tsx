import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";

import { EUserRole } from "@application/shared/enums";

function View({ role }: Props) {
    const roleMap: Record<EUserRole, string> = {
        [EUserRole.Admin]: "Admin",
        [EUserRole.Member]: "Member",
    };

    const roleColorMap: Record<EUserRole, string> = {
        [EUserRole.Admin]: "bg-primary text-white",
        [EUserRole.Member]: "bg-purple-500 text-white",
    };

    return (
        <Badge className={cn(roleColorMap[role] || "bg-primary text-primary-foreground", "h-5")}>
            {roleMap[role] || role}
        </Badge>
    );
}

interface Props {
    role: EUserRole;
}

export const UserRoleBadge = memo(View);
