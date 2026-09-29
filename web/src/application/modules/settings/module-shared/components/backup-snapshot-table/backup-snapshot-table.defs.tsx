import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { ArchiveRestore, Copy, Info, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import {
    canRestoreSnapshot,
    isRepoActive,
    snapshotScopeModuleId,
    sourceLabel,
    tagClassName,
    userTags,
} from "./backup-snapshot-table.helpers";

export interface BackupSnapshotRowActions {
    onViewDetails: (snapshot: BackupSnapshot) => void;
    onRestore: (snapshot: BackupSnapshot) => void;
    onDelete: (snapshot: BackupSnapshot) => void;
}

function MenuCell({ scope, snapshot, actions }: { scope: BackupSnapshotScope } & MenuProps) {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem
                    onClick={() => {
                        actions.onViewDetails(snapshot);
                    }}
                >
                    <Info className="mr-2 size-4" />
                    View details
                </DropdownMenuItem>
                <DropdownMenuItem
                    onClick={() => {
                        void navigator.clipboard.writeText(snapshot.snapshotId).then(() => {
                            toast.success("Snapshot ID copied");
                        });
                    }}
                >
                    <Copy className="mr-2 size-4" />
                    Copy ID
                </DropdownMenuItem>
                {canRestoreSnapshot(snapshot) && (
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                        triggerClassName="w-full"
                    >
                        {({ isDenied }) => (
                            <DropdownMenuItem
                                disabled={isDenied || !isRepoActive(snapshot)}
                                title={isRepoActive(snapshot) ? undefined : "The snapshot's repository is not active"}
                                onClick={() => {
                                    actions.onRestore(snapshot);
                                }}
                            >
                                <ArchiveRestore className="mr-2 size-4" />
                                Restore
                            </DropdownMenuItem>
                        )}
                    </PermissionTooltipAction>
                )}
                <PermissionTooltipAction
                    id={snapshotScopeModuleId(scope)}
                    action="delete"
                    triggerClassName="w-full"
                >
                    {({ isDenied }) => (
                        <DropdownMenuItem
                            disabled={isDenied}
                            className="text-destructive"
                            onClick={() => {
                                actions.onDelete(snapshot);
                            }}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Delete
                        </DropdownMenuItem>
                    )}
                </PermissionTooltipAction>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface MenuProps {
    snapshot: BackupSnapshot;
    actions: BackupSnapshotRowActions;
}

function createColumns(scope: BackupSnapshotScope, actions: BackupSnapshotRowActions): ColumnDef<BackupSnapshot>[] {
    const columns: ColumnDef<BackupSnapshot>[] = [
        {
            id: "time",
            header: "Taken",
            cell: ({ row: { original } }) => (
                <span className="whitespace-nowrap">{format(original.time, "yyyy-MM-dd HH:mm")}</span>
            ),
        },
        {
            id: "shortId",
            header: "Snapshot",
            cell: ({ row: { original } }) => <span className="font-mono">{original.shortId}</span>,
        },
        {
            id: "repo",
            header: "Repository",
            cell: ({ row: { original } }) => original.repo.name || original.repo.id,
        },
    ];
    if (scope.type !== "app") {
        columns.push({
            id: "app",
            header: "App",
            cell: ({ row: { original } }) => {
                if (!original.app) {
                    return "-";
                }
                if (original.app.deleted) {
                    return <span className="text-muted-foreground">Deleted app</span>;
                }
                return `${original.app.name} (${original.app.env})`;
            },
        });
    }
    columns.push(
        {
            id: "job",
            header: "Job",
            cell: ({ row: { original } }) => {
                if (!original.job) {
                    return "-";
                }
                return original.job.deleted ? (
                    <span className="text-muted-foreground">Deleted job</span>
                ) : (
                    original.job.name
                );
            },
        },
        {
            id: "source",
            header: "Source",
            meta: { align: "center", titleAlign: "center" },
            cell: ({ row: { original } }) =>
                original.source ? <Badge variant="secondary">{sourceLabel(original.source)}</Badge> : "-",
        },
        {
            id: "size",
            header: "Size",
            meta: { align: "right", titleAlign: "right" },
            cell: ({ row: { original } }) => formatDataSizeCompact(original.sizeBytes),
        },
        {
            id: "tags",
            header: "Tags",
            cell: ({ row: { original } }) => (
                <div className="flex flex-wrap gap-1">
                    {userTags(original).map(tag => (
                        <Badge
                            key={tag}
                            variant="outline"
                            className={cn("font-mono text-[11px]", tagClassName(tag))}
                        >
                            {tag}
                        </Badge>
                    ))}
                </div>
            ),
        },
        {
            id: "actions",
            header: "",
            size: 56,
            meta: { align: "center", titleAlign: "center" },
            cell: ({ row: { original } }) => (
                <MenuCell
                    scope={scope}
                    snapshot={original}
                    actions={actions}
                />
            ),
        },
    );
    return columns;
}

export const BackupSnapshotTableDefs = Object.freeze({
    columns: createColumns,
});
