import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingSslCert } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";
import { formatSslCertType } from "~/settings/module-shared/constants/ssl-provider.constants";

import { ESslCertType } from "@application/shared/enums";

import { SslCertEditCell, SslCertMenuCell } from "./building-blocks";
import type { SslCertTableScope } from "./ssl-cert-table.types";

function getCertTypeClassName(certType: SettingSslCert["certType"]): string {
    switch (certType) {
        case ESslCertType.Custom:
        case ESslCertType.SelfSigned:
            return "bg-slate-300 text-white";
        default:
            return "bg-emerald-300 text-white";
    }
}

function createColumns(scope: SslCertTableScope): ColumnDef<SettingSslCert>[] {
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
                <SslCertEditCell
                    scope={scope}
                    id={original.id}
                    inherited={original.inherited}
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
            accessorKey: "domain",
            header: "Domain",
            enableSorting: true,
        },
        {
            accessorKey: "certType",
            header: "Type",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex justify-center">
                    <Badge className={getCertTypeClassName(original.certType)}>
                        {formatSslCertType(original.certType)}
                    </Badge>
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
                        <Badge className="bg-purple-500 text-white">Inherited</Badge>
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
                <SslCertMenuCell
                    scope={scope}
                    sslCert={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const SslCertTableDefs = Object.freeze({
    columns: createColumns,
});
