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
    },
    {
        accessorKey: "key",
        header: "Key",
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
            return <span className="text-muted-foreground  whitespace-pre-wrap line-clamp-2">{note || "-"}</span>;
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
        header: "Actions",
        cell: ({ row: { original } }) => {
            return <MenuCell project={original} />;
        },
    },
];

export const ProjectsTableDefs = Object.freeze({
    columns,
});
