import React, { type ReactNode } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { type ColumnDef } from "@tanstack/react-table";
import type { AppStorageMount } from "~/projects/domain";
import { EMountType } from "~/projects/module-shared/enums";

import { PopConfirm } from "@application/shared/components";

type StorageMountWithId = AppStorageMount & { _id: string };

const mountTypeColorMap: Record<EMountType, string> = {
    [EMountType.Bind]: "bg-cyan-500 text-white hover:bg-cyan-500/90",
    [EMountType.Volume]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [EMountType.Cluster]: "bg-indigo-400 text-white hover:bg-indigo-400/90",
    [EMountType.Tmpfs]: "bg-rose-400 text-white hover:bg-rose-400/90",
    [EMountType.Npipe]: "bg-muted text-foreground hover:bg-muted",
    [EMountType.Image]: "bg-muted text-foreground hover:bg-muted",
};

function getSourceDisplay(mount: AppStorageMount): ReactNode {
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
                options.push(`size=${mount.tmpfsOptions.size}`);
            }
            if (mount.tmpfsOptions?.mode) {
                options.push(`mode=${mount.tmpfsOptions.mode}`);
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
            cell: ({ row }) => {
                const { type } = row.original;
                if (!type) return <div className="font-medium">-</div>;

                return (
                    <Badge className={cn("justify-center rounded-full text-sm lowercase", mountTypeColorMap[type])}>
                        {type}
                    </Badge>
                );
            },
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
