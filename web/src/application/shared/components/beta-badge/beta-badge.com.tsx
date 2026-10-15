import { cn } from "@lib/utils";

import { Badge } from "@/components/ui/badge";

/**
 * Marks a feature that is being tried out: it works, and may still change.
 * Small, to sit beside a title.
 */
export function BetaBadge({ className }: { className?: string }) {
    return (
        <Badge
            tone="violet"
            className={cn("px-1.5 py-0 text-[10px] font-semibold uppercase tracking-wide", className)}
            title="Beta: being tried out. It works, and may still change."
        >
            Beta
        </Badge>
    );
}
