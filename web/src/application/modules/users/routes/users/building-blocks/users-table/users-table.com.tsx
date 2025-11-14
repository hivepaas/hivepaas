import type { ColumnDef } from "@tanstack/react-table";

import type { User } from "@application/shared/entities";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { DataTable } from "@/components/ui";

const mockUsers: User[] = [
    {
        id: "1",
        email: "test@test.com",
        role: EUserRole.Admin,
        status: "active",
        fullName: "Test User",
        photo: null,
        position: "Test Position",
        securityOption: ESecuritySettings.PasswordOnly,
        createdAt: new Date(),
        updatedAt: null,
        accessExpireAt: null,
        lastAccess: null,
    },
    {
        id: "2",
        email: "test2@test.com",
        role: EUserRole.Admin,
        status: "active",
        fullName: "Test User 2",
        photo: null,
        position: "Test Position 2",
        securityOption: ESecuritySettings.PasswordOnly,
        createdAt: new Date(),
        updatedAt: null,
        accessExpireAt: null,
        lastAccess: null,
    },
    {
        id: "3",
        email: "test3@test.com",
        role: EUserRole.Admin,
        status: "active",
        fullName: "Test User 3",
        photo: null,
        position: "Test Position 3",
        securityOption: ESecuritySettings.PasswordOnly,
        createdAt: new Date(),
        updatedAt: null,
        accessExpireAt: null,
        lastAccess: null,
    },
    {
        id: "4",
        email: "test4@test.com",
        role: EUserRole.Admin,
        status: "active",
        fullName: "Test User 4",
        photo: null,
        position: "Test Position 4",
        securityOption: ESecuritySettings.PasswordOnly,
        createdAt: new Date(),
        updatedAt: null,
        accessExpireAt: null,
        lastAccess: null,
    },
    {
        id: "5",
        email: "test5@test.com",
        role: EUserRole.Admin,
        status: "active",
        fullName: "Test User 5",
        photo: null,
        position: "Test Position 5",
        securityOption: ESecuritySettings.PasswordOnly,
        createdAt: new Date(),
        updatedAt: null,
        accessExpireAt: null,
        lastAccess: null,
    },
    {
        id: "6",
        email: "test6@test.com",
        role: EUserRole.Admin,
        status: "active",
        fullName: "Test User 6",
        photo: null,
        position: "Test Position 6",
        securityOption: ESecuritySettings.PasswordOnly,
        createdAt: new Date(),
        updatedAt: null,
        accessExpireAt: null,
        lastAccess: null,
    },
];

const formatRole = (role: EUserRole): string => {
    const roleMap: Record<EUserRole, string> = {
        [EUserRole.Owner]: "Owner",
        [EUserRole.Admin]: "Admin",
        [EUserRole.Member]: "Member",
    };
    return roleMap[role] || role;
};

const formatStatus = (status: User["status"]): string => {
    const statusMap: Record<User["status"], string> = {
        active: "Active",
        pending: "Pending",
        disabled: "Disabled",
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

const columns: ColumnDef<User>[] = [
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
    return (
        <DataTable
            columns={columns}
            data={mockUsers}
            pageSize={10}
            enablePagination
            enableSorting
            onPaginationChange={value => {
                console.log(value);
            }}
        />
    );
}
