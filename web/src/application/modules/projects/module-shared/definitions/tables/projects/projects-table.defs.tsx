import { Avatar } from "@components/ui/avatar";
import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectBaseEntity } from "~/projects/domain";
import { ProjectStatusBadge } from "~/projects/module-shared/components";

import { ActionsCell, MenuCell } from "./building-blocks";

const columns: ColumnDef<ProjectBaseEntity>[] = [
    {
        id: "actions",
        header: "",
        minSize: 80,
        size: 80,
        meta: {
            align: "center",
            titleAlign: "center",
        },
        cell: ({ row: { original } }) => {
            return <ActionsCell id={original.id} />;
        },
    },
    {
        accessorKey: "name",
        header: "Name",
        cell: ({ row: { original } }) => (
            <div className="flex items-center gap-2">
                <Avatar
                    name={original.name}
                    src={original.photo}
                    className="rounded-lg"
                />
                <span>{original.name}</span>
            </div>
        ),
    },
    {
        accessorKey: "key",
        header: "Key",
        cell: ({ row: { original } }) => (
            <Badge
                tone="neutral"
                className="font-mono"
            >
                {original.key}
            </Badge>
        ),
    },
    {
        header: "Status",
        cell: ({ row: { original } }) => {
            const { status } = original;
            return <ProjectStatusBadge status={status} />;
        },
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        accessorKey: "note",
        header: "Note",
        cell: ({ row: { original } }) => {
            const { note } = original;
            return (
                <span className="text-muted-foreground whitespace-pre-wrap break-all line-clamp-2">{note || "-"}</span>
            );
        },
    },
    {
        accessorKey: "updatedAt",
        header: "Last Updated",
        cell: ({ row: { original } }) => {
            const { updatedAt } = original;
            if (!updatedAt) return "-";
            try {
                return format(updatedAt, "yyyy-MM-dd HH:mm:ss");
            } catch {
                return "-";
            }
        },
    },
    {
        id: "menu",
        header: "",
        enableResizing: false,
        enableHiding: false,
        minSize: 40,
        size: 40,
        cell: ({ row: { original } }) => <MenuCell id={original.id} />,
        meta: {
            align: "center",
            titleAlign: "center",
            sticky: "right",
        },
    },
];

export const ProjectsTableDefs = Object.freeze({
    columns,
});
