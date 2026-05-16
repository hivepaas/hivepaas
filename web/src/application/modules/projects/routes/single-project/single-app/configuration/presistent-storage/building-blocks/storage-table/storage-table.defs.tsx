import React, { type ReactNode } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { type ColumnDef } from "@tanstack/react-table";
import { EyeIcon, MoreVertical, Trash2Icon } from "lucide-react";
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

    switch (mount.type) {
        case EMountType.Bind: {
            if (mount.bindOptions?.subpath) {
                options.push(`Subpath: ${mount.bindOptions.subpath}`);
            }
            break;
        }
        case EMountType.Volume:
        case EMountType.Cluster: {
            const opts = mount.type === EMountType.Volume ? mount.volumeOptions : (mount.clusterOptions ?? {});
            if (opts?.subpath) {
                options.push(`Subpath: ${opts.subpath}`);
            }
            break;
        }
        case EMountType.Tmpfs:
            if (mount.tmpfsOptions?.size) {
                options.push(`Size: ${mount.tmpfsOptions.size}`);
            }
            if (mount.tmpfsOptions?.mode) {
                options.push(`Mode: ${mount.tmpfsOptions.mode}`);
            }
            break;
    }

    return options.join("\n") || "-";
}

export function createStorageTableColumns(
    onEdit: (mount: StorageMountWithId) => void,
    onDelete: (mount: StorageMountWithId) => Promise<void> | void,
): ColumnDef<StorageMountWithId>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row }) => (
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-link hover:opacity-50"
                    onClick={() => {
                        onEdit(row.original);
                    }}
                >
                    <EyeIcon className="size-5" />
                    <span className="sr-only">Edit storage mount</span>
                </Button>
            ),
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "type",
            header: "Type",
            cell: ({ row }) => {
                const { type } = row.original;
                if (!type) return <div className="font-medium">-</div>;

                return <Badge className={cn(mountTypeColorMap[type])}>{type}</Badge>;
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
            cell: ({ row }) => (
                <div className="text-sm text-muted-foreground whitespace-pre-line wrap-break-word">
                    {getOptionsDisplay(row.original)}
                </div>
            ),
            meta: {
                align: "left",
            },
        },
        {
            id: "actions",
            header: "",
            cell: ({ row }) => (
                <DropdownMenu>
                    <DropdownMenuTrigger
                        asChild
                        className="h-8 w-8"
                    >
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                        >
                            <MoreVertical className="size-4" />
                            <span className="sr-only">Actions menu</span>
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <div className="flex flex-col gap-0">
                            <PopConfirm
                                title="Remove Storage Mount"
                                variant="destructive"
                                confirmText="Remove"
                                cancelText="Cancel"
                                description="Are you sure you want to remove this storage mount?"
                                onConfirm={() => {
                                    void onDelete(row.original);
                                }}
                            >
                                <Button
                                    className="justify-start py-1.5"
                                    variant="ghost"
                                >
                                    <Trash2Icon className="mr-2 size-4" />
                                    Remove
                                </Button>
                            </PopConfirm>
                        </div>
                    </DropdownMenuContent>
                </DropdownMenu>
            ),
            meta: {
                align: "right",
            },
        },
    ];
}
