import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingEmail } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EEmailKind } from "@application/shared/enums";

import { EmailAccountEditCell, EmailAccountMenuCell } from "./building-blocks";
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

function getKindTone(kind: SettingEmail["kind"]): BadgeTone {
    switch (kind) {
        case EEmailKind.SMTP:
            return "emerald";
        case EEmailKind.HTTP:
            return "neutral";
        default:
            return "neutral";
    }
}

function createColumns(scope: EmailAccountTableScope): ColumnDef<SettingEmail>[] {
    return [
        {
            id: "view",
            accessorKey: "inherited",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <EmailAccountEditCell
                    scope={scope}
                    id={original.id}
                />
            ),
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
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
                    <Badge tone={getKindTone(original.kind)}>{formatKind(original.kind)}</Badge>
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
                    {scope.type === "project" && original.inherited && <Badge tone="purple">Inherited</Badge>}
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
