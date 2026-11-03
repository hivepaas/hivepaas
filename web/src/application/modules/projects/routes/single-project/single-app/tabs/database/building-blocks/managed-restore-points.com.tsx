import { History, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { AppManagedCommands } from "~/projects/data";
import { useManagedRestoreDialog } from "~/projects/dialogs/managed-restore";
import { useManagedRestorePointDialog } from "~/projects/dialogs/managed-restore-point";
import { type AppManagedRestorePoint, MANAGED_RESTORE_TARGET } from "~/projects/domain";
import { formatManagedTime, managedEngineOf } from "~/projects/module-shared/utils";

import { PopConfirm } from "@application/shared/components";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface Props {
    projectId: string;
    env: string;
    appId: string;
    appName: string;
    /** The database's engine, which the copy names what it archives after. */
    engine: string;
    points: AppManagedRestorePoint[];
    canWrite: boolean;
    /** A restoring database has no archive of its own yet to mark. */
    canMark: boolean;
}

/**
 * The moments a person named, to come back to exactly. Each is restorable until
 * retention deletes the base backup before it, unless it is kept.
 */
export function ManagedRestorePoints({ projectId, env, appId, appName, engine, points, canWrite, canMark }: Props) {
    const target = { projectId, env, appId };
    const { actions: pointDialog } = useManagedRestorePointDialog();
    const { actions: restoreDialog } = useManagedRestoreDialog();
    const { mutate: deletePoint, isPending: isDeleting } = AppManagedCommands.useDeleteRestorePoint({
        onSuccess: (_, request) => {
            toast.success(`Restore point ${request.name} deleted`);
        },
    });
    // The newest first.
    const sorted = [...points].sort((a, b) => b.at.getTime() - a.at.getTime());

    return (
        <section
            aria-label="Restore points"
            className="flex flex-col gap-3"
        >
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex flex-col">
                    <h2 className="text-sm font-semibold">Restore points</h2>
                    <p className="text-xs text-muted-foreground">
                        Name a moment, before a migration or an import, to restore to it exactly.
                    </p>
                </div>
                <Button
                    size="sm"
                    onClick={() => {
                        pointDialog.openCreate(target);
                    }}
                    disabled={!canWrite || !canMark}
                >
                    <Plus className="size-3.5" />
                    Make a restore point
                </Button>
            </div>

            {sorted.length === 0 ? (
                <p className="rounded-md border border-dashed px-3 py-4 text-center text-xs text-muted-foreground">
                    No restore point yet. The CLI makes one too: hivepaas db mark NAME.
                </p>
            ) : (
                <div className="rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Made</TableHead>
                                <TableHead>Kept until</TableHead>
                                <TableHead>Restorable until</TableHead>
                                <TableHead className="w-0" />
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {sorted.map(point => (
                                <TableRow key={point.name}>
                                    <TableCell className="align-top">
                                        <div className="flex flex-col gap-0.5">
                                            <span className="inline-flex items-center gap-1.5 font-medium">
                                                {point.name}
                                                {point.origin === "deploy" && (
                                                    <Badge
                                                        tone="sky"
                                                        className="text-[10px]"
                                                        title="A deployment made it. It is removed once no base backup reaches it."
                                                    >
                                                        Deploy
                                                    </Badge>
                                                )}
                                            </span>
                                            {point.note !== "" && (
                                                <span className="text-xs text-muted-foreground whitespace-pre-wrap">
                                                    {point.note}
                                                </span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell className="align-top text-xs tabular-nums">
                                        {formatManagedTime(point.at)}
                                    </TableCell>
                                    <TableCell className="align-top text-xs tabular-nums">
                                        {point.keepUntil ? (
                                            <span className="inline-flex items-center gap-1.5">
                                                {formatManagedTime(point.keepUntil)}
                                                {point.kept && (
                                                    <Badge
                                                        tone="violet"
                                                        className="text-[10px]"
                                                    >
                                                        Holding back
                                                    </Badge>
                                                )}
                                            </span>
                                        ) : (
                                            <span className="text-muted-foreground">Not kept</span>
                                        )}
                                    </TableCell>
                                    <TableCell className="align-top text-xs tabular-nums">
                                        {point.restorableUntil ? (
                                            formatManagedTime(point.restorableUntil)
                                        ) : (
                                            <Badge
                                                tone="neutral"
                                                className="text-[10px]"
                                            >
                                                No longer restorable
                                            </Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="align-top">
                                        <div className="flex items-center justify-end gap-1">
                                            <Button
                                                variant="ghost"
                                                size="icon-sm"
                                                aria-label={`Restore to ${point.name}`}
                                                title="Restore to it"
                                                disabled={!canWrite || !point.restorableUntil}
                                                onClick={() => {
                                                    restoreDialog.open(
                                                        { ...target, appName },
                                                        { kind: MANAGED_RESTORE_TARGET.point, point: point.name },
                                                    );
                                                }}
                                            >
                                                <History className="size-3.5" />
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="icon-sm"
                                                aria-label={`Edit ${point.name}`}
                                                title="Edit"
                                                disabled={!canWrite}
                                                onClick={() => {
                                                    pointDialog.openEdit(target, point);
                                                }}
                                            >
                                                <Pencil className="size-3.5" />
                                            </Button>
                                            <PopConfirm
                                                title={`Delete ${point.name}?`}
                                                description={
                                                    "Only the record goes: the mark stays in the " +
                                                    `${managedEngineOf(engine).archiveNoun}, and a kept point stops ` +
                                                    "holding retention back."
                                                }
                                                confirmText="Delete"
                                                variant="destructive"
                                                onConfirm={() => {
                                                    deletePoint({
                                                        projectID: projectId,
                                                        env,
                                                        appID: appId,
                                                        name: point.name,
                                                    });
                                                }}
                                            >
                                                <Button
                                                    variant="ghost"
                                                    size="icon-sm"
                                                    aria-label={`Delete ${point.name}`}
                                                    title="Delete"
                                                    disabled={!canWrite || isDeleting}
                                                >
                                                    <Trash2 className="size-3.5" />
                                                </Button>
                                            </PopConfirm>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            )}
        </section>
    );
}
