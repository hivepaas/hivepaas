import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { Button } from "@components/ui/button";
import { format } from "date-fns";
import { ArchiveRestore } from "lucide-react";
import type { BackupSnapshot } from "~/settings/domain";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";

import { isRepoActive, sourceLabel, tagClassName } from "./backup-snapshot-table.helpers";

function Row({ label, children }: { label: string; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1">
            <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{label}</span>
            <div className="text-sm text-foreground break-all">{children}</div>
        </div>
    );
}

/** A snapshot's details: what the table has no room for. */
export function BackupSnapshotDetails({ snapshot, runLink, onRestore, onOpenChange }: Props) {
    return (
        <Sheet
            open={Boolean(snapshot)}
            onOpenChange={onOpenChange}
        >
            <SheetContent className="sm:max-w-[520px] overflow-y-auto">
                {snapshot && (
                    <>
                        <SheetHeader>
                            <SheetTitle>Snapshot {snapshot.shortId}</SheetTitle>
                            <SheetDescription>{format(snapshot.time, "yyyy-MM-dd HH:mm:ss")}</SheetDescription>
                        </SheetHeader>
                        <div className="px-4">
                            <PermissionTooltipAction
                                id={MODULE_IDS.Project}
                                action="write"
                            >
                                {({ isDenied }) => (
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        disabled={isDenied || !isRepoActive(snapshot)}
                                        onClick={() => {
                                            onRestore(snapshot);
                                        }}
                                    >
                                        <ArchiveRestore className="size-4" />
                                        Restore
                                    </Button>
                                )}
                            </PermissionTooltipAction>
                        </div>
                        <div className="flex flex-col gap-4 px-4 pb-6">
                            <Row label="Snapshot ID">
                                <span className="font-mono select-all">{snapshot.snapshotId}</span>
                            </Row>
                            <Row label="Repository">{snapshot.repo.name || snapshot.repo.id}</Row>
                            <Row label="Size">{formatDataSizeCompact(snapshot.sizeBytes)}</Row>
                            <Row label="Source">{sourceLabel(snapshot.source)}</Row>
                            {snapshot.app && (
                                <Row label="App">
                                    {snapshot.app.deleted
                                        ? `Deleted app ${snapshot.app.id}`
                                        : `${snapshot.app.name} (${snapshot.app.env})`}
                                </Row>
                            )}
                            {snapshot.job && (
                                <Row label="Job">
                                    {snapshot.job.deleted ? `Deleted job ${snapshot.job.id}` : snapshot.job.name}
                                </Row>
                            )}
                            {snapshot.runId && <Row label="Run">{runLink ?? snapshot.runId}</Row>}
                            {snapshot.description && <Row label="Description">{snapshot.description}</Row>}
                            <Row label="Paths">
                                <span className="font-mono">{snapshot.paths.join(", ") || "-"}</span>
                            </Row>
                            {snapshot.hostname && <Row label="Hostname">{snapshot.hostname}</Row>}
                            <Row label="Tags">
                                <div className="flex flex-wrap gap-1.5">
                                    {snapshot.tags.length === 0 && "-"}
                                    {snapshot.tags.map(tag => (
                                        <Badge
                                            key={tag}
                                            variant="outline"
                                            className={cn("font-mono text-[11px]", tagClassName(tag))}
                                        >
                                            {tag}
                                        </Badge>
                                    ))}
                                </div>
                            </Row>
                        </div>
                    </>
                )}
            </SheetContent>
        </Sheet>
    );
}

interface Props {
    snapshot: BackupSnapshot | null;
    /** A link to the run that took the snapshot, when the view can make one. */
    runLink?: React.ReactNode;
    /** Opens the snapshot's restore. */
    onRestore: (snapshot: BackupSnapshot) => void;
    onOpenChange: (open: boolean) => void;
}
