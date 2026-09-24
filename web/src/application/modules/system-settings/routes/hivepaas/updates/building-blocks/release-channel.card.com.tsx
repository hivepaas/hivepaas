import { cn } from "@lib/utils";
import { ExternalLink } from "lucide-react";
import type { PublishedRelease, ReleaseChannel } from "~/system-settings/domain";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { channelLabel, formatReleaseDate } from "./update-labels";

interface Props {
    channel: ReleaseChannel;
    release: PublishedRelease | null;
    canManage: boolean;
    onReview: (version: string) => void;
}

const DESCRIPTIONS: Record<ReleaseChannel, string> = {
    stable: "Tested releases, for installations people rely on. Recommended.",
    beta:
        "What comes next, before it is stable. Once on beta, you move back to stable only when a stable " +
        "release passes the version you run.",
};

function status(channel: ReleaseChannel, release: PublishedRelease) {
    switch (release.relation) {
        case "newer":
            return channel === "stable"
                ? {
                      label: "Update available",
                      className: "bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-300",
                  }
                : {
                      label: "Pre-release",
                      className: "bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-300",
                  };
        case "same":
            return { label: "You run this", className: "" };
        case "older":
            return { label: "Older than yours", className: "" };
        default:
            return { label: "", className: "" };
    }
}

/** One channel's latest release, against the one running. */
export function ReleaseChannelCard({ channel, release, canManage, onReview }: Props) {
    const current = release ? status(channel, release) : null;
    const recommended = channel === "stable" && release?.canUpdate === true;

    return (
        <section
            aria-labelledby={`${channel}-release-title`}
            className={cn(
                "flex flex-col gap-4 rounded-xl border bg-card p-5",
                recommended && "border-primary border-[1.5px]",
            )}
        >
            <div className="flex items-center justify-between gap-2">
                <h2
                    id={`${channel}-release-title`}
                    className="text-[15px] font-semibold"
                >
                    {channelLabel(channel)}
                </h2>
                {current?.label && (
                    <Badge
                        variant={current.className ? undefined : "secondary"}
                        className={cn("rounded-full", current.className)}
                    >
                        {current.label}
                    </Badge>
                )}
            </div>

            {release ? (
                <>
                    <div className="flex flex-wrap items-baseline gap-2.5">
                        <span className="font-mono text-3xl font-bold">{release.appVersion}</span>
                        {release.releaseDate && (
                            <span className="text-[13px] text-muted-foreground">
                                released {formatReleaseDate(release.releaseDate)}
                            </span>
                        )}
                    </div>
                    <p className="text-sm leading-relaxed text-muted-foreground">
                        {release.relation === "older"
                            ? "Moving to it would go back from what you run, which the updater does not do: " +
                              "database changes only run forward."
                            : DESCRIPTIONS[channel]}
                    </p>
                    <div className="mt-auto flex flex-wrap items-center gap-2">
                        {release.canUpdate && canManage && (
                            <Button
                                variant={recommended ? "default" : "outline"}
                                onClick={() => {
                                    onReview(release.appVersion);
                                }}
                            >
                                Review update
                            </Button>
                        )}
                        {release.notesUrl && (
                            <a
                                href={release.notesUrl}
                                target="_blank"
                                rel="noreferrer"
                                className="inline-flex items-center gap-1.5 px-2 text-[13px] font-medium underline-offset-4 hover:underline"
                            >
                                Release notes
                                <ExternalLink className="size-3.5" />
                            </a>
                        )}
                    </div>
                </>
            ) : (
                <p className="text-sm text-muted-foreground">Nothing is published on this channel.</p>
            )}
        </section>
    );
}
