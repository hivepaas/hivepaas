import type { ColumnDef } from "@tanstack/react-table";

import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

import { UsersQueries } from "@application/modules/users/data/queries";
import type { UserBase } from "@application/modules/users/domain";

import { DataTable } from "@/components/ui";

const formatRole = (role: EUserRole): string => {
    const roleMap: Record<EUserRole, string> = {
        [EUserRole.Owner]: "Owner",
        [EUserRole.Admin]: "Admin",
        [EUserRole.Member]: "Member",
    };
    return roleMap[role] || role;
};

const formatStatus = (status: EUserStatus): string => {
    const statusMap: Record<EUserStatus, string> = {
        [EUserStatus.Active]: "Active",
        [EUserStatus.Pending]: "Pending",
        [EUserStatus.Disabled]: "Disabled",
    };
    return statusMap[status] || status;
};

const formatSecurityOption = (option: ESecuritySettings): string => {
    const optionMap: Record<ESecuritySettings, string> = {
        [ESecuritySettings.PasswordOnly]: "Password Only",
        [ESecuritySettings.Password2FA]: "Password + 2FA",
        [ESecuritySettings.EnforceSSO]: "Enforce SSO",
    };
    return optionMap[option] || option;
};

const formatDate = (date: Date | null): string => {
    if (!date) return "-";
    return new Intl.DateTimeFormat("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    }).format(date);
};

const columns: ColumnDef<UserBase>[] = [
    {
        accessorKey: "fullName",
        header: "Full Name",
    },
    {
        accessorKey: "email",
        header: "Email",
    },
    {
        accessorKey: "role",
        header: "Role",
        cell: ({ row }) => formatRole(row.original.role),
    },
    {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => formatStatus(row.original.status),
    },
    {
        accessorKey: "position",
        header: "Position",
    },
    {
        // accessorKey: "securityOption",
        header: "Security",
        cell: ({ row }) => formatSecurityOption(row.original.securityOption),
    },
    {
        accessorKey: "createdAt",
        header: "Created At",
        cell: ({ row }) => formatDate(row.original.createdAt),
    },
    {
        accessorKey: "lastAccess",
        header: "Last Access",
        cell: ({ row }) => formatDate(row.original.lastAccess),
    },
];

export function UsersTable() {
    const { data: { data: users } = DEFAULT_PAGINATED_DATA, isLoading } = UsersQueries.useFindManyPaginated();

    return (
        <DataTable
            columns={columns}
            data={users}
            pageSize={10}
            enablePagination
            enableSorting
            onPaginationChange={value => {
                console.log(value);
            }}
        />
    );
}
