import { memo } from "react";

import { Button } from "@components/ui";
import { Monitor, Network, ShieldCheck, Trash2 } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { NodesCommands } from "~/cluster/data/commands";
import { NodesQueries } from "~/cluster/data/queries";
import { NodeStatusBadge } from "~/cluster/module-shared/components";

import { BackButton } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { SingleNodeBreadcrumbs } from "../buidling-blocks";

import { SingleNodeHeaderSkeleton } from "./single-node-header.skeleton.com";

function View({ nodeId }: Props) {
    const { data, isLoading, error } = NodesQueries.useFindOneById({ id: nodeId });
    const { navigate } = useAppNavigate();

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
        deleteOne(
            { id: node.id },
            {
                onSuccess: () => {
                    toast.success("Node removed successfully");
                    navigate.modules(ROUTE.cluster.nodes.$route);
                },
            },
        );
    };

    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <SingleNodeBreadcrumbs node={node} />
                <div className="flex items-center gap-2">
                    <PopConfirm
                        title="Remove node"
                        variant="destructive"
                        confirmText="Remove"
                        cancelText="Cancel"
                        description="Are you sure you want to remove this node?"
                        onConfirm={handleRemove}
                    >
                        <Button
                            variant="outline"
                            disabled={isDeleting}
                        >
                            <Trash2 className="mr-2 size-4" />
                            Remove
                        </Button>
                    </PopConfirm>
                </div>
            </div>

            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex items-center gap-4">
                    <div className="flex flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">{node.name}</h2>
                            <NodeStatusBadge status={node.status} />
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                                <Monitor className="size-4 text-blue-500" />
                                <div className="flex gap-1">
                                    <span>Hostname:</span>
                                    <span className="text-foreground">{node.hostname}</span>
                                </div>
                            </div>
                            <span>•</span>
                            <div className="flex items-center gap-1.5 text-sm">
                                <Network className="size-4 text-blue-500" />
                                <div className="flex gap-1">
                                    <span>IP:</span>
                                    <span className="text-foreground">{node.addr}</span>
                                </div>
                            </div>
                            <span>•</span>
                            <div className="flex items-center gap-1.5 text-sm">
                                <ShieldCheck className="size-4 text-blue-500" />
                                <div className="flex gap-1">
                                    <span>Role:</span>
                                    <span className="text-foreground">{node.isLeader ? "Leader" : "Member"}</span>
                                </div>
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
