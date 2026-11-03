import { cn } from "@/lib/utils";
import type { VulnSummary } from "~/projects/api/services";

import { SEVERITIES } from "./severity";

const TILE_TONES: Record<string, string> = {
    critical: "border-red-500/30 bg-red-500/5 text-red-700 dark:text-red-400",
    high: "border-orange-500/30 bg-orange-500/5 text-orange-700 dark:text-orange-400",
    medium: "border-amber-500/30 bg-amber-500/5 text-amber-700 dark:text-amber-400",
    low: "border-sky-500/30 bg-sky-500/5 text-sky-700 dark:text-sky-400",
    unknown: "border-slate-500/30 bg-slate-500/5 text-slate-700 dark:text-slate-400",
};

/** The findings by severity, each with how many have a fixed version. */
export function SeverityCounters({ summary }: { summary: VulnSummary }) {
    return (
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
            {SEVERITIES.map(({ key, label }) => {
                const count = summary[key];
                return (
                    <div
                        key={key}
                        data-testid={`severity-${key}`}
                        className={cn("flex flex-col gap-0.5 rounded-lg border px-3 py-2.5", TILE_TONES[key])}
                    >
                        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
                        <span className="text-2xl font-semibold tabular-nums text-foreground">{count.total}</span>
                        <span className="text-xs text-muted-foreground">{count.fixable} fixable</span>
                    </div>
                );
            })}
        </div>
    );
}
