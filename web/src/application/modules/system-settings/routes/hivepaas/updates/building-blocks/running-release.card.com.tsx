import type { RunningRelease } from "~/system-settings/domain";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { channelLabel, formatReleaseDate } from "./update-labels";

interface Props {
    current: RunningRelease | null;
    checkedAt: number;
    isChecking: boolean;
    onCheck: () => void;
}

/** The release this installation runs, and when the published ones were last looked at. */
export function RunningReleaseCard({ current, checkedAt, isChecking, onCheck }: Props) {
    const released = formatReleaseDate(current?.releaseDate ?? null);

    return (
        <section
            aria-labelledby="running-release-title"
            className="flex flex-wrap items-center justify-between gap-4 rounded-xl border bg-card px-5 py-4"
        >
            <div className="flex flex-col gap-1.5">
                <h2
                    id="running-release-title"
                    className="text-[13px] font-medium text-muted-foreground"
                >
                    Running now
                </h2>
                <div className="flex flex-wrap items-center gap-2.5">
                    <span className="font-mono text-2xl font-bold">{current?.appVersion ?? "–"}</span>
                    {current && <Badge variant="secondary">{channelLabel(current.channel)}</Badge>}
                    {released && <span className="text-[13px] text-muted-foreground">released {released}</span>}
                </div>
            </div>
            <div className="flex items-center gap-3">
                {checkedAt > 0 && (
                    <span className="text-[13px] text-muted-foreground">
                        Checked at {new Date(checkedAt).toLocaleTimeString()}
                    </span>
                )}
                <Button
                    variant="outline"
                    size="sm"
                    isLoading={isChecking}
                    onClick={onCheck}
                >
                    Check now
                </Button>
            </div>
        </section>
    );
}
