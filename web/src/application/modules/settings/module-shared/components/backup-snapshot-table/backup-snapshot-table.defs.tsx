import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { BackupSnapshotMenuCell } from "./backup-snapshot-menu-cell.com";
import { sourceLabel, tagTone, userTags } from "./backup-snapshot-table.helpers";

export interface BackupSnapshotRowActions {
    onViewDetails: (snapshot: BackupSnapshot) => void;
    onRestore: (snapshot: BackupSnapshot) => void;
    onDelete: (snapshot: BackupSnapshot) => void;
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
                return (
                    <Badge tone="cyan">
                        {original.app.name} ({original.app.env})
                    </Badge>
                );
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
                            tone={tagTone(tag)}
                            className="font-mono text-[11px]"
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
                <BackupSnapshotMenuCell
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
