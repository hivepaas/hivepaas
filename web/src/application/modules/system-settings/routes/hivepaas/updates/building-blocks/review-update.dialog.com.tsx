import { useEffect, useState } from "react";

import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { cn } from "@lib/utils";
import { AlertTriangle, ArrowRight, Clock, ExternalLink, FileDown } from "lucide-react";
import { HivePaaSUpdatesCommands, HivePaaSUpdatesQueries } from "~/system-settings/data";
import type { HivePaaSReleaseInfo, ReleaseChannel, UpdateComponent } from "~/system-settings/domain";

import { ROUTE } from "@application/shared/constants";

import { Button, Checkbox, Separator } from "@/components/ui";
import { Skeleton } from "@/components/ui/skeleton";

import { channelLabel, componentName, formatReleaseDate, imageTag } from "./update-labels";

interface Props {
    open: boolean;
    /** The channel the operator chose to review; the other can still be picked here. */
    channel: ReleaseChannel;
    releaseInfo: HivePaaSReleaseInfo;
    onOpenChange: (open: boolean) => void;
    /** The update has been started, and HivePaaS is about to stop. */
    onStarted: (version: string) => void;
}

const CHANGE_TAGS: Record<string, { label: string; className: string }> = {
    update: { label: "Updated", className: "bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300" },
    major: { label: "Major", className: "bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-300" },
    blocked: { label: "Blocked", className: "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300" },
};

function ComponentRow({ component }: { component: UpdateComponent }) {
    const tag = CHANGE_TAGS[component.change];
    const blocked = component.change === "blocked";

    return (
        <li className="grid grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_minmax(0,1fr)_96px] items-center gap-x-3 gap-y-1 border-b border-border/60 px-4 py-2.5 text-[13px] last:border-b-0">
            <span className="font-semibold">{componentName(component)}</span>
            <span
                className="truncate font-mono text-muted-foreground"
                title={component.currentImage}
            >
                {imageTag(component.currentImage)}
            </span>
            <span
                className="truncate font-mono"
                title={component.targetImage}
            >
                {imageTag(component.targetImage)}
            </span>
            {tag && (
                <span
                    className={cn("justify-self-end rounded-md px-2 py-0.5 text-[11px] font-semibold", tag.className)}
                >
                    {tag.label}
                </span>
            )}
            {blocked && <span className="col-span-full text-xs text-destructive">{component.reason}</span>}
            {component.requiresBackup && (
                <span className="col-span-full text-xs text-amber-800 dark:text-amber-300">
                    The new database is loaded from the backup, so the backup cannot be skipped.
                </span>
            )}
            {component.interruptsTraffic && (
                <span className="col-span-full text-xs text-amber-800 dark:text-amber-300">
                    Sites served through HivePaaS are unreachable for a few seconds while it restarts.
                </span>
            )}
        </li>
    );
}

/**
 * What an update would do, read from the services as they run now, and the
 * choices it leaves: which channel, and whether to back the database up first.
 * Components cannot be picked one by one: a release is tried as a whole.
 */
export function ReviewUpdateDialog({ open, channel, releaseInfo, onOpenChange, onStarted }: Props) {
    const offered = (["stable", "beta"] as const).filter(c => releaseInfo[c]?.canUpdate === true);
    const [selected, setSelected] = useState<ReleaseChannel>(channel);
    const [backup, setBackup] = useState(true);
    const [configSaved, setConfigSaved] = useState(false);

    useEffect(() => {
        if (open) {
            setSelected(channel);
            setBackup(true);
            setConfigSaved(false);
        }
    }, [open, channel]);

    const version = releaseInfo[selected]?.appVersion ?? "";
    const { data, isLoading } = HivePaaSUpdatesQueries.useFindPlan({ targetVersion: version }, { enabled: open });
    const plan = data?.data;
    const { mutate: update, isPending } = HivePaaSUpdatesCommands.useUpdate({
        onSuccess: () => {
            onStarted(version);
        },
    });

    const changed = plan?.components.filter(c => c.change !== "none" && c.change !== "not-deployed") ?? [];
    const unchanged = plan?.components.filter(c => c.change === "none").map(componentName) ?? [];
    const mustBackUp = plan?.requiresBackup === true;
    const released = formatReleaseDate(plan?.target.releaseDate ?? null);

    return (
        <Dialog
            open={open}
            onOpenChange={next => {
                if (!isPending) {
                    onOpenChange(next);
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[760px]">
                <DialogHeader>
                    <DialogTitle>Update HivePaaS</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-5">
                    <div className="flex flex-wrap items-center gap-2.5 text-sm">
                        <span className="font-mono text-muted-foreground">{releaseInfo.current?.appVersion}</span>
                        <ArrowRight className="size-4 text-muted-foreground" />
                        <span className="font-mono font-bold">{version}</span>
                        <span className="rounded-full bg-secondary px-2 py-0.5 text-xs font-semibold">
                            {channelLabel(selected)}
                        </span>
                        {released && <span className="text-muted-foreground">released {released}</span>}
                        {plan?.target.notesUrl && (
                            <a
                                href={plan.target.notesUrl}
                                target="_blank"
                                rel="noreferrer"
                                className="inline-flex items-center gap-1 font-medium underline-offset-4 hover:underline"
                            >
                                Release notes
                                <ExternalLink className="size-3.5" />
                            </a>
                        )}
                    </div>

                    {offered.length > 1 && (
                        <fieldset className="flex flex-col gap-2">
                            <legend className="mb-2 text-[13px] font-semibold">Update to</legend>
                            <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
                                {offered.map(c => (
                                    <label
                                        key={c}
                                        className={cn(
                                            "flex cursor-pointer items-center gap-2.5 rounded-lg border px-3.5 py-2.5 text-[13px]",
                                            selected === c && "border-primary border-[1.5px]",
                                        )}
                                    >
                                        <input
                                            type="radio"
                                            name="update-channel"
                                            checked={selected === c}
                                            disabled={isPending}
                                            onChange={() => {
                                                setSelected(c);
                                            }}
                                        />
                                        <span>
                                            <b>{channelLabel(c)}</b>{" "}
                                            <span className="font-mono">{releaseInfo[c]?.appVersion}</span>
                                            {c === "stable" && (
                                                <span className="text-muted-foreground"> · recommended</span>
                                            )}
                                        </span>
                                    </label>
                                ))}
                            </div>
                        </fieldset>
                    )}

                    <section
                        aria-labelledby="update-changes-title"
                        className="rounded-lg border"
                    >
                        <div className="flex flex-wrap items-center justify-between gap-2 border-b px-4 py-3">
                            <h3
                                id="update-changes-title"
                                className="text-sm font-semibold"
                            >
                                What changes
                            </h3>
                            {unchanged.length > 0 && (
                                <span className="text-xs text-muted-foreground">Unchanged: {unchanged.join(", ")}</span>
                            )}
                        </div>
                        {isLoading || !plan ? (
                            <div className="flex flex-col gap-2 p-4">
                                <Skeleton className="h-5 w-full" />
                                <Skeleton className="h-5 w-full" />
                                <Skeleton className="h-5 w-2/3" />
                            </div>
                        ) : changed.length === 0 ? (
                            <p className="px-4 py-4 text-sm text-muted-foreground">
                                Every component already runs what this release names.
                            </p>
                        ) : (
                            <>
                                <div className="grid grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_minmax(0,1fr)_96px] gap-x-3 border-b px-4 py-2 text-xs font-medium text-muted-foreground">
                                    <span>Component</span>
                                    <span>Now</span>
                                    <span>After</span>
                                    <span />
                                </div>
                                <ul>
                                    {changed.map(component => (
                                        <ComponentRow
                                            key={component.key}
                                            component={component}
                                        />
                                    ))}
                                </ul>
                            </>
                        )}
                    </section>

                    {plan?.blocked ? (
                        <div className="flex items-start gap-2.5 rounded-lg border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive">
                            <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                            <span>
                                This update cannot run here: a component would cross a major version the release does
                                not allow. It needs to be migrated by hand first.
                            </span>
                        </div>
                    ) : (
                        <div className="flex items-start gap-2.5 rounded-lg border bg-muted/40 px-3.5 py-2.5 text-sm">
                            <Clock className="mt-0.5 size-4 shrink-0" />
                            <span>
                                The dashboard and the API stop for <b>a few minutes</b> while HivePaaS updates itself.
                                Your apps keep running.
                            </span>
                        </div>
                    )}

                    <label
                        htmlFor="update-backup"
                        className="flex items-start gap-2.5 text-sm leading-6"
                    >
                        <Checkbox
                            id="update-backup"
                            className="mt-1"
                            checked={backup || mustBackUp}
                            disabled={isPending || mustBackUp}
                            onCheckedChange={checked => {
                                setBackup(checked === true);
                            }}
                        />
                        <span>
                            <span className="block font-semibold">Back up the HivePaaS database first</span>
                            <span className="block text-xs leading-normal text-muted-foreground">
                                {mustBackUp
                                    ? "Required by this update: the new database is loaded from it."
                                    : "A failed migration is undone from this backup. Turn it off only when the backup " +
                                      "itself cannot be made, such as on a full disk."}
                            </span>
                        </span>
                    </label>

                    <div className="flex flex-col gap-3 rounded-lg border border-amber-300 bg-amber-50 px-3.5 py-3 text-sm dark:border-amber-800 dark:bg-amber-950/60">
                        <div className="flex items-start gap-2.5">
                            <FileDown className="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-300" />
                            <span className="leading-6">
                                Export the system configuration before updating. The database backup undoes a failed
                                migration; the exported spec is what the projects, apps and settings can be set up again
                                from if anything else goes wrong.{" "}
                                <a
                                    href={ROUTE.operations.export.$route}
                                    target="_blank"
                                    rel="noreferrer"
                                    className="inline-flex items-center gap-1 font-medium underline underline-offset-4"
                                >
                                    Open Operations › Export
                                    <ExternalLink className="size-3.5" />
                                </a>
                            </span>
                        </div>
                        <label
                            htmlFor="update-config-saved"
                            className="flex items-start gap-2.5 font-medium"
                        >
                            <Checkbox
                                id="update-config-saved"
                                className="mt-0.5"
                                checked={configSaved}
                                disabled={isPending}
                                onCheckedChange={checked => {
                                    setConfigSaved(checked === true);
                                }}
                            />
                            I confirm I have backed up the system configuration
                        </label>
                    </div>
                </DialogBody>
                <DialogActionFooter>
                    <Button
                        variant="outline"
                        disabled={isPending}
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={!plan || plan.blocked || !configSaved || isPending}
                        isLoading={isPending}
                        onClick={() => {
                            update({ targetVersion: version, skipBackup: !(backup || mustBackUp) });
                        }}
                    >
                        Update to {version}
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
