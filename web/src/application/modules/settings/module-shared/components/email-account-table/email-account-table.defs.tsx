import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingEmail } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EEmailKind } from "@application/shared/enums";

import { EmailAccountMenuCell } from "./building-blocks";
import type { EmailAccountTableScope } from "./email-account-table.types";

function formatKind(kind: SettingEmail["kind"]): string {
    switch (kind) {
        case EEmailKind.SMTP:
            return "SMTP";
        case EEmailKind.HTTP:
            return "HTTP";
        default:
            return kind;
    }
}

function getKindClassName(kind: SettingEmail["kind"]): string {
    switch (kind) {
        case EEmailKind.SMTP:
            return "bg-emerald-300 text-white";
        case EEmailKind.HTTP:
            return "bg-slate-300 text-white";
        default:
            return "bg-muted text-muted-foreground";
    }
}

function createColumns(scope: EmailAccountTableScope): ColumnDef<SettingEmail>[] {
    return [
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            id: "username",
            header: "Username",
            cell: ({ row: { original } }) => original.smtp?.username ?? original.http?.username ?? "-",
        },
        {
            accessorKey: "kind",
            header: "Type",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex justify-center">
                    <span
                        className={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-medium ${getKindClassName(
                            original.kind,
                        )}`}
                    >
                        {formatKind(original.kind)}
                    </span>
                </div>
            ),
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <SettingStatusBadge status={original.status} />
                    {scope.type === "project" && original.inherited && (
                        <span className="inline-flex items-center rounded-md bg-purple-500 px-2 py-1 text-xs font-medium text-white">
                            Inherited
                        </span>
                    )}
                </div>
            ),
        },
        {
            accessorKey: "expireAt",
            header: "Expire At",
            cell: ({ row: { original } }) => {
                if (!original.expireAt) {
                    return "-";
                }

                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <EmailAccountMenuCell
                    scope={scope}
                    emailAccount={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const EmailAccountTableDefs = Object.freeze({
    columns: createColumns,
});
