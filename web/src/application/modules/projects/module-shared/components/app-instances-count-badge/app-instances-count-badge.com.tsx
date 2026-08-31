import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import {
    APP_REPLICAS_STATUS_BADGE_CLASS,
    getAppReplicasStatusLabel,
    resolveAppReplicasStatus,
} from "~/projects/module-shared/utils";

function View({ current, total }: Props) {
    const status = resolveAppReplicasStatus(current, total);
    const label = getAppReplicasStatusLabel(current, total);

    return (
        <Badge
            className={cn(
                // Pill shape sets this counter apart from the square-ish badges used elsewhere.
                "h-5 rounded-full px-2 text-xs font-medium leading-none",
                APP_REPLICAS_STATUS_BADGE_CLASS[status],
            )}
            title={label}
            aria-label={label}
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
