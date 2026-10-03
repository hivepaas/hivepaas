import type { LoggingPerformanceNode, LoggingPerformanceStatusReason } from "~/system-settings/domain";

import { obiPreflightReasonText } from "@application/shared/constants";

import { Badge } from "@/components/ui/badge";

import { capacityLabel } from "./performance-texts";

/**
 * What a node's agent last said of it: whether it can run OBI, and if it does, for how many apps. Nothing
 * when the statuses were not read; "no status" when its agent said nothing in the last few minutes.
 */
export function PerformanceNodeStatus({
    node,
    statusReason,
}: {
    node: LoggingPerformanceNode;
    statusReason: LoggingPerformanceStatusReason | null;
}) {
    if (statusReason) {
        return <span className="text-muted-foreground">-</span>;
    }
    const { status } = node;
    if (!status) {
        return (
            <span
                className="text-muted-foreground"
                title="Its agent said nothing in the last 3 minutes: one from before this version of HivePaaS, or one starting."
            >
                No status yet
            </span>
        );
    }
    if (!status.ok) {
        return (
            <span className="flex flex-col gap-1">
                <Badge
                    tone="red"
                    className="w-fit"
                >
                    Cannot run OBI
                </Badge>
                <span className="text-xs text-muted-foreground">
                    {status.reasons.map(obiPreflightReasonText).join("; ")}
                </span>
            </span>
        );
    }
    const capacity = capacityLabel(status.capacity);
    if (status.running) {
        return (
            <span className="flex flex-col gap-1">
                <Badge
                    tone="green"
                    className="w-fit"
                >
                    Running
                </Badge>
                <span className="text-xs text-muted-foreground">
                    {status.apps === 1 ? "1 app" : `${status.apps} apps`}, {capacity}
                </span>
            </span>
        );
    }
    return (
        <span className="flex flex-col gap-1">
            <Badge
                tone="neutral"
                className="w-fit"
            >
                {status.wanted ? "Not running" : "Can run OBI"}
            </Badge>
            <span className="text-xs text-muted-foreground">
                {!status.wanted
                    ? `Kernel ${status.kernel || "unknown"}`
                    : status.apps === 0
                      ? "No app running here asks for its routes and calls."
                      : "Starting, or it could not start: the HivePaaS agent's logs on this node say why."}
            </span>
        </span>
    );
}
