import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { ArchiveRestore, Copy, Info, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

import type { BackupSnapshotRowActions } from "./backup-snapshot-table.defs";
import { canRestoreSnapshot, isRepoActive, snapshotScopeModuleId } from "./backup-snapshot-table.helpers";

/** A snapshot's row menu: its details, its ID, a restore, a deletion. */
export function BackupSnapshotMenuCell({ scope, snapshot, actions }: { scope: BackupSnapshotScope } & MenuProps) {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
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
                <DropdownMenuItem
                    onClick={() => {
                        actions.onViewDetails(snapshot);
                    }}
                >
                    <Info className="mr-2 size-4" />
                    View details
                </DropdownMenuItem>
                <DropdownMenuItem
                    onClick={() => {
                        void navigator.clipboard.writeText(snapshot.snapshotId).then(() => {
                            toast.success("Snapshot ID copied");
                        });
                    }}
                >
                    <Copy className="mr-2 size-4" />
                    Copy ID
                </DropdownMenuItem>
                {canRestoreSnapshot(snapshot) && (
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                        triggerClassName="w-full"
                    >
                        {({ isDenied }) => (
                            <DropdownMenuItem
                                disabled={isDenied || !isRepoActive(snapshot)}
                                title={isRepoActive(snapshot) ? undefined : "The snapshot's repository is not active"}
                                onClick={() => {
                                    actions.onRestore(snapshot);
                                }}
                            >
                                <ArchiveRestore className="mr-2 size-4" />
                                Restore
                            </DropdownMenuItem>
                        )}
                    </PermissionTooltipAction>
                )}
                <PermissionTooltipAction
                    id={snapshotScopeModuleId(scope)}
                    action="delete"
                    triggerClassName="w-full"
                >
                    {({ isDenied }) => (
                        <DropdownMenuItem
                            disabled={isDenied || !isRepoActive(snapshot)}
                            title={isRepoActive(snapshot) ? undefined : "The snapshot's repository is not active"}
                            className="text-destructive"
                            onClick={() => {
                                actions.onDelete(snapshot);
                            }}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Delete
                        </DropdownMenuItem>
                    )}
                </PermissionTooltipAction>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface MenuProps {
    snapshot: BackupSnapshot;
    actions: BackupSnapshotRowActions;
}
