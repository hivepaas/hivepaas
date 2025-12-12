import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { ENodeAvailability } from "~/cluster/module-shared/enums";

function View({ availability }: Props) {
    const availabilityMap: Record<ENodeAvailability, string> = {
        [ENodeAvailability.Active]: "Active",
        [ENodeAvailability.Pause]: "Pause",
        [ENodeAvailability.Drain]: "Drain",
    };

    const availabilityColorMap: Record<ENodeAvailability, string> = {
        [ENodeAvailability.Active]: "bg-green-500 text-white",
        [ENodeAvailability.Pause]: "bg-purple-500 text-white",
        [ENodeAvailability.Drain]: "bg-orange-500 text-white",
    };

    return (
        <Badge className={cn(availabilityColorMap[availability] || "bg-primary text-primary-foreground", "h-6")}>
            {availabilityMap[availability] || availability}
        </Badge>
    );
}

interface Props {
    availability: ENodeAvailability;
}

export const NodeAvailabilityBadge = memo(View);
