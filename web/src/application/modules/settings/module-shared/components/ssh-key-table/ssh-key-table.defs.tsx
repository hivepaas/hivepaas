import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingSSHKey } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { ESSHKeyKind } from "@application/shared/enums";

import { SSHKeyEditCell, SSHKeyMenuCell } from "./building-blocks";
import type { SSHKeyTableScope } from "./ssh-key-table.types";

const SSH_KEY_KIND_BADGE_TONE: BadgeTone = "sky";

function formatKeyType(keyType?: string) {
    if (keyType === undefined || keyType === "") {
        return "Unspecified";
    }

    return keyType;
}

/** The color of a key's algorithm; one not named, or not known, is neutral. */
function keyTypeTone(keyType?: string): BadgeTone {
    switch (keyType?.toLowerCase()) {
        case "ed25519":
            return "emerald";
        case "rsa":
            return "amber";
        case "ecdsa":
            return "blue";
        default:
            return "neutral";
    }
}

function createColumns(scope: SSHKeyTableScope): ColumnDef<SettingSSHKey>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <SSHKeyEditCell
                    scope={scope}
                    id={original.id}
                />
            ),
            meta: { align: "center", titleAlign: "center" },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            accessorKey: "kind",
            header: "Type",
            // A key stored with no kind is a plain git key: the form shows it as one.
            cell: ({ row: { original } }) => (
                <Badge tone={SSH_KEY_KIND_BADGE_TONE}>{original.kind ? original.kind : ESSHKeyKind.Git}</Badge>
            ),
        },
        {
            accessorKey: "keyType",
            header: "Key Type",
            cell: ({ row: { original } }) => (
                <Badge
                    tone={keyTypeTone(original.keyType)}
                    className="font-mono"
                >
                    {formatKeyType(original.keyType)}
                </Badge>
            ),
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: { align: "center", titleAlign: "center" },
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
                if (!original.expireAt) return "-";
                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <SSHKeyMenuCell
                    scope={scope}
                    sshKey={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const SSHKeyTableDefs = Object.freeze({
    columns: createColumns,
});
