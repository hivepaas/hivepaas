import { memo } from "react";

import { Avatar, Button } from "@components/ui";
import { Monitor, Network, ShieldCheck, Trash2 } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { NodesCommands } from "~/cluster/data/commands";
import { NodesQueries } from "~/cluster/data/queries";
import { NodeStateBadge } from "~/cluster/module-shared/components";

import { PopConfirm } from "@application/shared/components/pop-confirm";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui/separator";

import { SingleNodeBreadcrumbs } from "../buidling-blocks";

import { SingleNodeHeaderSkeleton } from "./single-node-header.skeleton.com";

function View({ nodeId }: Props) {
    const { data, isLoading, error } = NodesQueries.useFindOneById({ id: nodeId });
    const { navigate } = useAppNavigate();
    const { canDelete } = useConditionalModule({ id: MODULE_IDS.Cluster });

    const { mutate: deleteOne, isPending: isDeleting } = NodesCommands.useDeleteOne({});

    if (isLoading) {
        return <SingleNodeHeaderSkeleton />;
    }

    if (error) {
        return null;
    }

    invariant(data, "data must be defined");

    const { data: node } = data;

    const handleRemove = () => {
        if (!canDelete) {
            return;
        }

        deleteOne(
            { id: node.id, force: false },
            {
                onSuccess: () => {
                    toast.success("Node removed successfully");
                    navigate.modules(ROUTE.cluster.nodes.$route);
                },
            },
        );
    };

    return (
        <div className="bg-background pt-3 sm:pt-4 px-3 sm:px-5 rounded-lg">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <SingleNodeBreadcrumbs node={node} />
                <div className="flex items-center gap-2">
                    {canDelete ? (
                        <PopConfirm
                            title="Remove node"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Confirm deletion of this item?"
                            onConfirm={handleRemove}
                        >
                            <Button
                                variant="outline"
                                size="sm"
                                className="h-8 text-xs sm:text-sm"
                                disabled={isDeleting}
                            >
                                <Trash2 className="mr-1.5 size-3.5 sm:size-4" />
                                Remove
                            </Button>
                        </PopConfirm>
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.Cluster}
                            action="delete"
                        >
                            {({ isDenied }) => (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 text-xs sm:text-sm"
                                    disabled={isDenied}
                                >
                                    <Trash2 className="mr-1.5 size-3.5 sm:size-4" />
                                    Remove
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                </div>
            </div>

            <Separator className="my-2.5 sm:my-3 opacity-50" />

            <div className="flex items-start gap-3 sm:gap-4 pb-3 sm:pb-4 min-w-0">
                <Avatar
                    name={node.name || "<unset>"}
                    className="size-9 sm:size-12 text-sm sm:text-lg shrink-0 rounded-xl"
                />
                <div className="flex flex-1 min-w-0 flex-col gap-1 sm:gap-1.5 justify-center">
                    <div className="flex flex-wrap items-center gap-2">
                        <h2 className="text-lg sm:text-xl font-semibold text-foreground truncate">
                            {node.name || "<unset>"}
                        </h2>
                        <NodeStateBadge state={node.state} />
                    </div>
                    <div className="flex flex-wrap items-center gap-x-3 sm:gap-x-4 gap-y-1 text-xs sm:text-sm text-muted-foreground">
                        <div className="flex items-center gap-1.5 min-w-0">
                            <Monitor className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 min-w-0">
                                <span className="shrink-0">Hostname:</span>
                                <span className="text-foreground font-medium truncate">{node.hostname || "-"}</span>
                            </div>
                        </div>
                        <span className="text-amber-500/70 dark:text-amber-400/70 select-none hidden sm:inline">•</span>
                        <div className="flex items-center gap-1.5 min-w-0">
                            <Network className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 min-w-0">
                                <span className="shrink-0">IP:</span>
                                <span className="text-foreground font-medium truncate">{node.addr || "-"}</span>
                            </div>
                        </div>
                        <span className="text-amber-500/70 dark:text-amber-400/70 select-none hidden sm:inline">•</span>
                        <div className="flex items-center gap-1.5 min-w-0">
                            <ShieldCheck className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 min-w-0">
                                <span className="shrink-0">Role:</span>
                                <span className="text-foreground font-medium">
                                    {node.isLeader ? "Leader" : "Member"}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

interface Props {
    nodeId: string;
}

export const SingleNodeHeader = memo(View);
