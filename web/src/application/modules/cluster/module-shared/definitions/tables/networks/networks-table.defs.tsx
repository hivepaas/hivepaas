import { Checkbox } from "@components/ui";
import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import { ActionsCell, MenuCell } from "./building-blocks";

function BooleanCell({ value }: { value: boolean }) {
    return (
        <Checkbox
            checked={value}
            disabled
            aria-label={value ? "Enabled" : "Disabled"}
        />
    );
}

function createColumns(scope: NetworkManagementScope): ColumnDef<ClusterNetwork>[] {
    return [
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
                return (
                    <ActionsCell
                        network={original}
                        scope={scope}
                    />
                );
            },
        },
        {
            accessorKey: "name",
            header: "Name",
        },
        ...(scope.type === "project"
            ? [
                  {
                      id: "inheritedStatus",
                      header: "Status",
                      meta: {
                          align: "center",
                          titleAlign: "center",
                      },
                      cell: ({ row: { original } }) => {
                          if (!original.inherited) {
                              return "-";
                          }

                          return <Badge className="bg-purple-500 text-white">Inherited</Badge>;
                      },
                  } satisfies ColumnDef<ClusterNetwork>,
              ]
            : []),
        {
            accessorKey: "driver",
            header: "Driver",
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "attachable",
            header: "Attachable",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => <BooleanCell value={original.attachable} />,
        },
        {
            accessorKey: "internal",
            header: "Internal",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => <BooleanCell value={original.internal} />,
        },
        {
            accessorKey: "enableIPv4",
            header: "IPv4",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => <BooleanCell value={original.enableIPv4} />,
        },
        {
            accessorKey: "enableIPv6",
            header: "IPv6",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => <BooleanCell value={original.enableIPv6} />,
        },
        {
            accessorKey: "createdAt",
            header: "Created At",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => format(original.createdAt, "yyyy-MM-dd HH:mm:ss"),
        },
        {
            header: "Actions",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => {
                return (
                    <MenuCell
                        network={original}
                        scope={scope}
                    />
                );
            },
        },
    ];
}

export const NetworksTableDefs = Object.freeze({
    columns: createColumns,
});
