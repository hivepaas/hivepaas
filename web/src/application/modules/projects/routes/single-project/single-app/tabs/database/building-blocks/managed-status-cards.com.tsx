import type React from "react";

import { cn } from "@lib/utils";
import { Lightbulb } from "lucide-react";
import { type AppManagedSettings, type AppManagedStatus, MANAGED_ARCHIVING_STATE } from "~/projects/domain";
import { formatManagedTime } from "~/projects/module-shared/utils";

import { AppLink } from "@application/shared/components";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Badge, type StatusMeaning, statusClassName } from "@/components/ui/badge";

import { PLACEMENT_COPY, SUGGESTION_COPY } from "./managed-format";

interface Props {
    settings: AppManagedSettings;
    status: AppManagedStatus;
    /** The app's backup snapshots, where its dumps are listed and restored from. */
    snapshotsHref: string;
}

/**
 * What a managed database's engine knows of it: whether its WAL is archived,
 * how far back a restore can go, its last base backup, restore test and dump,
 * and where its archive is from it.
 */
export function ManagedStatusCards({ settings, status, snapshotsHref }: Props) {
    const { archiving, restoreWindow, lastBackup, lastRestoreTest, lastDump } = status;
    const dumping = settings.dump.repo !== undefined;
    const placement = status.placement === "" ? undefined : PLACEMENT_COPY[status.placement];

    return (
        <div className="flex flex-col gap-3">
            <div className="grid grid-cols-1 gap-3 @xl:grid-cols-2 @4xl:grid-cols-4">
                <StatusCard
                    title="Archiving"
                    badge={archivingBadge(archiving.state, archiving.stale)}
                >
                    {archiving.lastArchivedAt ? (
                        <Line
                            label="Last archived"
                            title={formatManagedTime(archiving.lastArchivedAt)}
                        >
                            {timeAgoFormatter.format(archiving.lastArchivedAt)}
                        </Line>
                    ) : (
                        <Line label="Last archived">Nothing yet</Line>
                    )}
                    {archiving.failedCount > 0 && <Line label="Failed">{archiving.failedCount} times</Line>}
                    <Line label="WAL on the node">{formatDataSizeCompact(archiving.pgWalBytes)}</Line>
                    <Line label="Archive timeout">{settings.archiveTimeout}</Line>
                </StatusCard>

                <StatusCard title="Restore window">
                    {restoreWindow ? (
                        <>
                            <Line label="From">{formatManagedTime(restoreWindow.from)}</Line>
                            <Line label="To">{formatManagedTime(restoreWindow.to)}</Line>
                            <Line label="Base backups">
                                {status.backups.length} kept, of {settings.retainFull}
                            </Line>
                        </>
                    ) : (
                        <p className="text-xs text-muted-foreground">
                            None before the first base backup, at {settings.baseBackupSchedule}.
                        </p>
                    )}
                </StatusCard>

                <StatusCard
                    title="Last base backup"
                    badge={
                        lastBackup
                            ? resultBadge(
                                  lastBackup.succeeded ? "success" : "failure",
                                  lastBackup.succeeded,
                                  lastBackup.succeeded ? "Succeeded" : "Failed",
                              )
                            : null
                    }
                >
                    {lastBackup ? (
                        <>
                            <Line
                                label="Ended"
                                title={formatManagedTime(lastBackup.endedAt)}
                            >
                                {timeAgoFormatter.format(lastBackup.endedAt)}
                            </Line>
                            {lastBackup.bytes > 0 && (
                                <Line label="Size">{formatDataSizeCompact(lastBackup.bytes)}</Line>
                            )}
                            {lastBackup.name !== "" && (
                                <Line label="Name">
                                    <span className="font-mono text-[11px]">{lastBackup.name}</span>
                                </Line>
                            )}
                        </>
                    ) : (
                        <p className="text-xs text-muted-foreground">Not run yet: {settings.baseBackupSchedule}.</p>
                    )}
                </StatusCard>

                <StatusCard
                    title="Last restore test"
                    badge={
                        lastRestoreTest
                            ? resultBadge(lastRestoreTest.passed ? "success" : "failure", lastRestoreTest.passed)
                            : null
                    }
                >
                    {lastRestoreTest ? (
                        <>
                            <Line
                                label="Ended"
                                title={formatManagedTime(lastRestoreTest.endedAt)}
                            >
                                {timeAgoFormatter.format(lastRestoreTest.endedAt)}
                            </Line>
                            <Line label="Took">{formatSeconds(lastRestoreTest.durationSeconds)}</Line>
                            {lastRestoreTest.restoredTo !== "" && (
                                <Line label="Restored to">{lastRestoreTest.restoredTo}</Line>
                            )}
                        </>
                    ) : (
                        <p className="text-xs text-muted-foreground">Not run yet: {settings.restoreTestSchedule}.</p>
                    )}
                    <Line label="Runs in">
                        {settings.restoreTest?.mode === "verifier" ? "A verifier app" : "The database's container"}
                    </Line>
                </StatusCard>
            </div>

            {(dumping || lastDump) && (
                <section
                    aria-label="Last dump"
                    className="flex flex-wrap items-center gap-2 rounded-md border px-3 py-2 text-xs"
                >
                    <span className="text-muted-foreground">Last dump</span>
                    {lastDump ? (
                        <>
                            {resultBadge(
                                lastDump.succeeded ? "success" : "failure",
                                lastDump.succeeded,
                                lastDump.succeeded ? "Succeeded" : "Failed",
                            )}
                            <span title={formatManagedTime(lastDump.endedAt)}>
                                {timeAgoFormatter.format(lastDump.endedAt)}
                            </span>
                            {lastDump.succeeded && (
                                <span className="text-muted-foreground">{formatDataSizeCompact(lastDump.bytes)}</span>
                            )}
                        </>
                    ) : (
                        <span className="text-muted-foreground">Not run yet: {settings.dump.schedule}.</span>
                    )}
                    {!dumping && <span className="text-muted-foreground">Dumps are off.</span>}
                    <AppLink.Basic
                        to={snapshotsHref}
                        className="ml-auto underline"
                    >
                        Backup snapshots
                    </AppLink.Basic>
                </section>
            )}

            {placement && (
                <div className="flex flex-wrap items-center gap-2 rounded-md border px-3 py-2 text-xs">
                    <span className="text-muted-foreground">Archive placement</span>
                    <Badge tone={placement.tone}>{placement.label}</Badge>
                    <span className="text-muted-foreground">{placement.text}</span>
                    {status.heldBack && (
                        <span className="text-muted-foreground">
                            Kept points hold back {status.heldBack.backups} base{" "}
                            {status.heldBack.backups === 1 ? "backup" : "backups"},{" "}
                            {formatDataSizeCompact(status.heldBack.bytes)}.
                        </span>
                    )}
                </div>
            )}

            {status.suggestions.map(suggestion => (
                <div
                    key={suggestion}
                    role="note"
                    className="flex items-start gap-2 rounded-md border border-sky-500/30 bg-sky-500/10 px-3 py-2 text-xs text-sky-800 dark:text-sky-300"
                >
                    <Lightbulb className="mt-0.5 size-3.5 shrink-0" />
                    <span>{SUGGESTION_COPY[suggestion] ?? suggestion}</span>
                </div>
            ))}
        </div>
    );
}

function StatusCard({ title, badge, children }: { title: string; badge?: React.ReactNode; children: React.ReactNode }) {
    return (
        <section
            aria-label={title}
            className="flex flex-col gap-2 rounded-md border px-3 py-2.5"
        >
            <div className="flex items-center justify-between gap-2">
                <h3 className="text-xs font-medium text-muted-foreground">{title}</h3>
                {badge}
            </div>
            <div className="flex flex-col gap-1">{children}</div>
        </section>
    );
}

function Line({ label, title, children }: { label: string; title?: string; children: React.ReactNode }) {
    return (
        <div className="flex items-baseline justify-between gap-3 text-xs">
            <span className="text-muted-foreground">{label}</span>
            <span
                className="text-right tabular-nums"
                title={title}
            >
                {children}
            </span>
        </div>
    );
}

function archivingBadge(state: AppManagedStatus["archiving"]["state"], stale: boolean) {
    if (state === MANAGED_ARCHIVING_STATE.ok) {
        return resultBadge(stale ? "attention" : "success", true, stale ? "Behind" : "Archiving");
    }
    if (state === MANAGED_ARCHIVING_STATE.failing) {
        return resultBadge("failure", false, "Failing");
    }
    return resultBadge("inactive", false, "Unknown");
}

function resultBadge(meaning: StatusMeaning, ok: boolean, label?: string) {
    return <Badge className={cn("text-[11px]", statusClassName(meaning))}>{label ?? (ok ? "Passed" : "Failed")}</Badge>;
}

function formatSeconds(seconds: number): string {
    if (seconds < 60) {
        return `${seconds}s`;
    }
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
        return `${minutes}m ${seconds % 60}s`;
    }
    return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}
