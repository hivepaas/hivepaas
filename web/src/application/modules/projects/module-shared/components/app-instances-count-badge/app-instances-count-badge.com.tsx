import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";

function resolveClassName(current: number, total: number) {
    // All desired instances are up (0/0 included: nothing is expected, so nothing is wrong).
    if (current >= total) {
        return "bg-green-500 text-white";
    }

    // Partially up.
    if (current > 0) {
        return "bg-orange-500 text-white";
    }

    // Nothing is up while some instances are expected.
    return "bg-red-500 text-white";
}

function View({ current, total }: Props) {
    return (
        <Badge
            className={cn(
                // Pill shape sets this counter apart from the square-ish badges used elsewhere.
                "h-5 rounded-full px-2 text-xs font-medium leading-none",
                resolveClassName(current, total),
            )}
            aria-label={`${current} of ${total} instances running`}
        >
            {current}/{total}
        </Badge>
    );
}

interface Props {
    current: number;
    total: number;
}

export const AppInstancesCountBadge = memo(View);
