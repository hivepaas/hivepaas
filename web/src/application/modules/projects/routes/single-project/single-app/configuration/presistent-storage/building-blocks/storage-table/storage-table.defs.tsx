import React from "react";

import { type ColumnDef } from "@tanstack/react-table";
import type { AppStorageMount } from "~/projects/domain";
import { EMountType } from "~/projects/module-shared/enums";

import { PopConfirm } from "@application/shared/components";

type StorageMountWithId = AppStorageMount & { _id: string };

function getSourceDisplay(mount: AppStorageMount): string {
    switch (mount.type) {
        case EMountType.Bind:
            if (mount.bindOptions?.baseDir && mount.bindOptions.subpath) {
                return `${mount.bindOptions.baseDir}/${mount.bindOptions.subpath}`;
            }
            return mount.bindOptions?.baseDir ?? "";
        case EMountType.Volume:
            if (mount.volumeOptions?.volume && mount.volumeOptions.subpath) {
                return `${mount.volumeOptions.volume}:${mount.volumeOptions.subpath}`;
            }
            return mount.volumeOptions?.volume ?? "";
        case EMountType.Cluster:
            if (mount.clusterOptions?.volume && mount.clusterOptions.subpath) {
                return `${mount.clusterOptions.volume}:${mount.clusterOptions.subpath}`;
            }
            return mount.clusterOptions?.volume ?? "";
        case EMountType.Tmpfs:
            return "-";
        default:
            return "";
    }
}

function getOptionsDisplay(mount: AppStorageMount): string {
    const options: string[] = [];

    if (mount.readOnly) {
        options.push("ro");
    }
    if (mount.consistency) {
        options.push(`consistency=${mount.consistency}`);
    }

    switch (mount.type) {
        case EMountType.Bind:
            if (mount.bindOptions?.propagation) {
                options.push(`propagation=${mount.bindOptions.propagation}`);
            }
            break;
        case EMountType.Volume:
        case EMountType.Cluster: {
            const opts = mount.type === EMountType.Volume ? mount.volumeOptions : (mount.clusterOptions ?? {});
            if (opts?.noCopy) {
                options.push("nocopy");
            }
            if (opts?.labels) {
                const labelCount = Object.keys(opts.labels).length;
                if (labelCount > 0) {
                    options.push(`labels=${labelCount}`);
                }
            }
            break;
        }
        case EMountType.Tmpfs:
            if (mount.tmpfsOptions?.size) {
                options.push(`size=${mount.tmpfsOptions.size}B`);
            }
            if (mount.tmpfsOptions?.mode) {
                options.push(`mode=${mount.tmpfsOptions.mode.toString(8)}`);
            }
            break;
    }

    return options.join(", ") || "-";
}

export function createStorageTableColumns(
    onEdit: (mount: StorageMountWithId) => void,
    onDelete: (mount: StorageMountWithId) => void,
): ColumnDef<StorageMountWithId>[] {
    return [
        {
            accessorKey: "type",
            header: "Type",
            cell: ({ row }) => <div className="font-medium">{row.original.type ?? "-"}</div>,
            meta: {
                align: "left",
            },
        },
        {
            accessorKey: "source",
            header: "Source",
            cell: ({ row }) => <div className="text-sm">{getSourceDisplay(row.original)}</div>,
            meta: {
                align: "left",
            },
        },
        {
            accessorKey: "target",
            header: "Target",
            cell: ({ row }) => <div className="text-sm">{row.original.target ?? "-"}</div>,
            meta: {
                align: "left",
            },
        },
        {
            accessorKey: "options",
            header: "Options",
            cell: ({ row }) => <div className="text-sm text-muted-foreground">{getOptionsDisplay(row.original)}</div>,
            meta: {
                align: "left",
            },
        },
        {
            id: "actions",
            header: "",
            cell: ({ row }) => {
                return (
                    <div className="flex items-center justify-end gap-2">
                        <button
                            type="button"
                            onClick={() => {
                                onEdit(row.original);
                            }}
                            className="text-sm text-blue-600 hover:underline"
                        >
                            Edit
                        </button>
                        <PopConfirm
                            title="Remove Storage Mount"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Are you sure you want to remove this storage mount?"
                            onConfirm={() => {
                                onDelete(row.original);
                            }}
                        >
                            <button
                                type="button"
                                className="text-sm text-red-600 hover:underline"
                            >
                                Delete
                            </button>
                        </PopConfirm>
                    </div>
                );
            },
            meta: {
                align: "right",
            },
        },
    ];
}
