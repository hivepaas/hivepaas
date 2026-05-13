import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingSslCert } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { ESslCertType } from "@application/shared/enums";

import { SslCertMenuCell } from "./building-blocks";
import type { SslCertTableScope } from "./ssl-cert-table.types";

function formatCertType(certType: SettingSslCert["certType"]): string {
    switch (certType) {
        case ESslCertType.LetsEncrypt:
            return "Let’s Encrypt";
        case ESslCertType.Custom:
            return "custom";
        case ESslCertType.SelfSigned:
            return "self-signed";
        default:
            return certType;
    }
}

function getCertTypeClassName(certType: SettingSslCert["certType"]): string {
    switch (certType) {
        case ESslCertType.LetsEncrypt:
            return "bg-emerald-300 text-white";
        case ESslCertType.Custom:
            return "bg-slate-300 text-white";
        default:
            return "bg-muted text-muted-foreground";
    }
}

function createColumns(scope: SslCertTableScope): ColumnDef<SettingSslCert>[] {
    return [
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
                    <span
                        className={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-medium ${getCertTypeClassName(
                            original.certType,
                        )}`}
                    >
                        {formatCertType(original.certType)}
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
