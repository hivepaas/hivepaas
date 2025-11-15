import { Avatar } from "@components/ui/avatar";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { UserRoleBadge, UserSecurityBadge, UserStatusBadge } from "~/users/module-shared/components";

import type { UserBase } from "@application/modules/users/domain";

import { ActionsCell, UserMenuCell } from "./building-blocks";

const columns: ColumnDef<UserBase>[] = [
    {
        id: "actions",
        header: "Actions",
        enableResizing: false,
        enableHiding: false,
        minSize: 80,
        size: 80,
        cell: ({ row }) => <ActionsCell id={row.original.id} />,
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        accessorKey: "fullName",
        header: "Full Name",
        cell: ({ row }) => (
            <div className="flex items-center gap-2">
                <Avatar
                    name={row.original.fullName}
                    src={row.original.photo}
                />
                <span>{row.original.fullName}</span>
            </div>
        ),
    },
    {
        accessorKey: "email",
        header: "Email",
    },
    {
        header: "Role",
        cell: ({ row }) => <UserRoleBadge role={row.original.role} />,
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        header: "Status",
        cell: ({ row }) => <UserStatusBadge status={row.original.status} />,
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        header: "Position",
        cell: ({ row }) => <span>{row.original.position}</span>,
    },
    {
        // accessorKey: "securityOption",
        header: "Security",
        cell: ({ row }) => <UserSecurityBadge securityOption={row.original.securityOption} />,
    },
    {
        accessorKey: "createdAt",
        header: "Created At",
        cell: ({ row }) => format(row.original.createdAt, "yyyy-MM-dd HH:mm:ss"),
    },
    {
        accessorKey: "lastAccess",
        header: "Last Access",
        cell: ({ row }) => (row.original.lastAccess ? format(row.original.lastAccess, "yyyy-MM-dd HH:mm:ss") : "-"),
    },
    {
        id: "userMenu",
        header: "",
        enableResizing: false,
        enableHiding: false,
        minSize: 40,
        size: 40,
        cell: ({ row }) => (
            <UserMenuCell
                status={row.original.status}
                onActivate={() => {
                    // TODO: Implement activate user
                    console.log("Activate user:", row.original.id);
                }}
                onDisable={() => {
                    // TODO: Implement disable user
                    console.log("Disable user:", row.original.id);
                }}
            />
        ),
        meta: {
            align: "center",
            titleAlign: "center",
            sticky: "right",
        },
    },
];

export const UsersTableDefs = Object.freeze({
    columns,
});
