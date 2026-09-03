import React, { type ReactNode } from "react";

import { Badge } from "@components/ui/badge";
import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { type ColumnDef } from "@tanstack/react-table";
import { EyeIcon, MoreVertical, Trash2Icon } from "lucide-react";
import type { AppStorageMount } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

type StorageMountWithId = AppStorageMount & { _id: string };

const MOUNT_TYPE_BADGE_CLASS = "bg-cyan-500 text-white";

function getSourceDisplay(mount: AppStorageMount): ReactNode {
    return mount.source ?? "-";
}

function getOptionsDisplay(mount: AppStorageMount): string {
    const options: string[] = [];
    const volumeOpts = mount.volumeOptions ?? mount.clusterOptions;

    if (volumeOpts?.subpath) {
        options.push(`Subpath: ${volumeOpts.subpath}`);
    }

    if (volumeOpts?.noCopy) {
        options.push("No Copy");
    }

    if (mount.readOnly) {
        options.push("Read-only");
    }

    if (mount.type === "tmpfs") {
        if (mount.tmpfsOptions?.size) {
            options.push(`Size: ${mount.tmpfsOptions.size}`);
        }
        if (mount.tmpfsOptions?.mode) {
            options.push(`Mode: ${mount.tmpfsOptions.mode}`);
        }
    }

    return options.join("\n") || "-";
}

export function createStorageTableColumns(
    onEdit: (mount: StorageMountWithId) => void,
    onDelete: (mount: StorageMountWithId) => Promise<void> | void,
    canWrite: boolean,
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

                return <Badge className={MOUNT_TYPE_BADGE_CLASS}>{type}</Badge>;
            },
            meta: {
                align: "left",
            },
        },
        {
            accessorKey: "source",
            header: "Source",
            cell: ({ row }) => <div className="text-sm break-all">{getSourceDisplay(row.original)}</div>,
            meta: {
                align: "left",
            },
        },
        {
            accessorKey: "target",
            header: "Target",
            cell: ({ row }) => <div className="text-sm break-all">{row.original.target ?? "-"}</div>,
            meta: {
                align: "left",
            },
        },
        {
            accessorKey: "options",
            header: "Options",
            cell: ({ row }) => (
                <div className="text-sm text-muted-foreground whitespace-pre-line break-all">
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
                            {canWrite ? (
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
                            ) : (
                                <PermissionTooltipAction
                                    id={MODULE_IDS.Project}
                                    action="write"
                                    triggerClassName="w-full"
                                >
                                    {({ isDenied }) => (
                                        <Button
                                            className="justify-start py-1.5 w-full"
                                            variant="ghost"
                                            disabled={isDenied}
                                        >
                                            <Trash2Icon className="mr-2 size-4" />
                                            Remove
                                        </Button>
                                    )}
                                </PermissionTooltipAction>
                            )}
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
