import { Button } from "@components/ui";
import { toast } from "sonner";
import { ClusterNetworksCommands } from "~/cluster/data/commands";
import { EClusterNetworkDriver } from "~/cluster/module-shared/enums";
import { ProjectNetworksCommands } from "~/projects/data/commands";

import { RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { CreateNetworkForm } from "../../../dialogs/create-network/form";
import type { CreateNetworkFormOutput } from "../../../dialogs/create-network/schemas";
import { CLUSTER_NETWORK_FORM_ROUTE_PANEL_CLASS } from "../../constants/network-form-layout.constants";

type CreateNetworkScope = { type: "cluster" } | { type: "project"; projectId: string };

function toRecord(items: { key: string; value: string }[]) {
    return items.reduce<Record<string, string>>((acc, item) => {
        const key = item.key.trim();
        if (!key) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, {});
}

export function CreateNetworkFormRoute({ scope }: Props) {
    const { navigate } = useAppNavigate();
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const canWrite = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;

    function navigateToList() {
        navigate.modules(getNetworkListRoute(scope), { ignorePrevPath: true });
    }

    const { mutate: createClusterNetwork, isPending: isCreatingClusterNetwork } = ClusterNetworksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Network created");
            navigateToList();
        },
    });
    const { mutate: createProjectNetwork, isPending: isCreatingProjectNetwork } = ProjectNetworksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Network created");
            navigateToList();
        },
    });

    const isPending = isCreatingClusterNetwork || isCreatingProjectNetwork;

    function onSubmit(values: CreateNetworkFormOutput) {
        if (!canWrite) {
            return;
        }

        const payload = {
            name: values.name,
            driver: EClusterNetworkDriver.Overlay,
            enableIPv4: values.enableIPv4,
            enableIPv6: values.enableIPv6,
            internal: values.internal,
            attachable: values.attachable,
            ingress: values.ingress,
            labels: toRecord(values.labels),
            options: toRecord(values.options),
        };

        if (scope.type === "cluster") {
            createClusterNetwork({
                payload: {
                    ...payload,
                    availableInProjects: values.availableInProjects,
                },
            });
            return;
        }

        createProjectNetwork({
            projectID: scope.projectId,
            payload,
        });
    }

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <RouteFormHeader title="Create a network" />

            <div className={CLUSTER_NETWORK_FORM_ROUTE_PANEL_CLASS}>
                <CreateNetworkForm
                    readOnly={!canWrite}
                    isPending={isPending}
                    showAvailableInProjects={scope.type !== "project"}
                    showProjectNamePrefixNote={scope.type === "project"}
                    onSubmit={onSubmit}
                >
                    {!canWrite ? (
                        <div className="shrink-0 px-0 mt-6 flex justify-end">
                            <Button
                                type="button"
                                onClick={navigateToList}
                            >
                                Close
                            </Button>
                        </div>
                    ) : (
                        <div className="flex justify-end mt-6">
                            <div className="flex items-center gap-3">
                                <Button
                                    type="button"
                                    variant="outline"
                                    className="min-w-[100px]"
                                    disabled={isPending}
                                    onClick={navigateToList}
                                >
                                    Cancel
                                </Button>
                                <Button
                                    type="submit"
                                    disabled={isPending}
                                    className="min-w-[120px]"
                                >
                                    Save
                                </Button>
                            </div>
                        </div>
                    )}
                </CreateNetworkForm>
            </div>
        </div>
    );
}

function getNetworkListRoute(scope: CreateNetworkScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.networks.$route(scope.projectId);
    }

    return ROUTE.cluster.networks.$route;
}

interface Props {
    scope: CreateNetworkScope;
}
