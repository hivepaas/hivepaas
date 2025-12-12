import type { ColumnDef } from "@tanstack/react-table";
import type { NodeDetails } from "~/cluster/domain";
import { NodeAvailabilityBadge, NodeRoleBadge } from "~/cluster/module-shared/components";

import { ActionsCell, MenuCell } from "./building-blocks";

const columns: ColumnDef<NodeDetails>[] = [
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
        accessorKey: "hostname",
        header: "Hostname",
    },
    {
        accessorKey: "addr",
        header: "Address",
    },
    {
        header: "CPUs",
        cell: ({ row: { original } }) => {
            const { resources } = original;
            if (!resources) return "-";
            return <span>{resources.cpus} CPU</span>;
        },
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        header: "Memory",
        cell: ({ row: { original } }) => {
            const { resources } = original;
            if (!resources) return "-";
            return `${resources.memoryMB} MB`;
        },
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        header: "Role",
        meta: {
            align: "center",
            titleAlign: "center",
        },
        cell: ({ row: { original } }) => {
            const { role, isLeader } = original;
            return (
                <NodeRoleBadge
                    role={role}
                    isLeader={isLeader}
                />
            );
        },
    },
    {
        header: "Availability",
        cell: ({ row: { original } }) => {
            const { availability } = original;
            return <NodeAvailabilityBadge availability={availability} />;
        },
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        header: "Actions",
        cell: ({ row: { original } }) => {
            return <MenuCell node={original} />;
        },
    },
];

export const NodesTableDefs = Object.freeze({
    columns,
});
