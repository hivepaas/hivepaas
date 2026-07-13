import { Button } from "@components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { ClusterNetworksQueries } from "~/cluster/data/queries";
import { ProjectNetworksQueries } from "~/projects/data/queries";

import { ViewNetworkForm, type ViewNetworkFormOutput } from "../form";
import { useViewNetworkDialogState } from "../hooks";

export function ViewNetworkDialog() {
    const { state, props: dialogOptions, ...actions } = useViewNetworkDialogState();
    const open = state.mode !== "closed";
    const scope = state.mode === "open" ? state.scope : null;
    const targetNetwork = state.mode === "open" ? state.network : null;
    const networkID = targetNetwork?.id ?? "";
    const isInheritedProjectNetwork = scope?.type === "project" && targetNetwork?.inherited === true;

    const clusterNetworkQuery = ClusterNetworksQueries.useFindOneById(
        { networkID },
        {
            enabled: open && scope?.type === "cluster" && Boolean(networkID),
        },
    );
    const projectNetworkQuery = ProjectNetworksQueries.useFindOneById(
        {
            projectID: scope?.type === "project" ? scope.projectId : "",
            networkID,
        },
        {
            enabled: open && scope?.type === "project" && !isInheritedProjectNetwork && Boolean(networkID),
        },
    );
    const network = isInheritedProjectNetwork
        ? targetNetwork
        : scope?.type === "cluster"
          ? clusterNetworkQuery.data?.data
          : projectNetworkQuery.data?.data;
    const isFetching = scope?.type === "cluster" ? clusterNetworkQuery.isFetching : projectNetworkQuery.isFetching;

    function handleClose(): void {
        actions.close();
        dialogOptions?.onClose?.();
    }

    function handleReadonlySubmit(_values: ViewNetworkFormOutput) {
        void _values;
    }

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[820px]">
                <DialogHeader>
                    <DialogTitle className="text-2xl">Network info</DialogTitle>
                </DialogHeader>
                <DialogBody>
                    {network ? (
                        <ViewNetworkForm
                            key={network.id}
                            network={network}
                            readOnlyAvailableInProjects
                            readOnlyDefault
                            showAvailableInProjects={scope?.type === "cluster"}
                            onSubmit={handleReadonlySubmit}
                        />
                    ) : (
                        <div className="py-10 text-center text-sm text-muted-foreground">
                            {isFetching ? "Loading network..." : "Network not found"}
                        </div>
                    )}
                </DialogBody>
                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        className="min-w-[100px]"
                        onClick={handleClose}
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
