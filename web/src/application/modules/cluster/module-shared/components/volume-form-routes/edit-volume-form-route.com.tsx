import { Button } from "@components/ui";
import { toast } from "sonner";
import { ClusterVolumesCommands } from "~/cluster/data/commands";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";
import { ProjectClusterVolumesCommands } from "~/projects/data/commands";
import { ProjectClusterVolumesQueries } from "~/projects/data/queries";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { CreateOrEditVolumeForm, type CreateOrEditVolumeFormOutput } from "../volume-form";

import { toVolumeFormInitialValues } from "./volume-form-route.helpers";

export function EditVolumeFormRoute({ scope, volumeId }: Props) {
    const { navigate } = useAppNavigate();
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const canWrite = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;

    function navigateToList() {
        navigate.modules(getVolumeListRoute(scope), { ignorePrevPath: true });
    }

    const clusterVolumeQuery = ClusterVolumesQueries.useFindOneById(
        { volumeID: volumeId },
        {
            enabled: scope.type === "cluster" && Boolean(volumeId),
        },
    );
    const projectVolumeQuery = ProjectClusterVolumesQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            volumeID: volumeId,
        },
        {
            enabled: scope.type === "project" && Boolean(volumeId),
        },
    );
    const volume = scope.type === "cluster" ? clusterVolumeQuery.data?.data : projectVolumeQuery.data?.data;
    const isFetching = scope.type === "cluster" ? clusterVolumeQuery.isFetching : projectVolumeQuery.isFetching;
    const isInherited = scope.type === "project" && volume?.inherited === true;
    const canSubmit = canWrite && !isInherited;

    const { mutate: updateClusterVolume, isPending: isUpdatingClusterVolume } = ClusterVolumesCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Volume updated");
            navigateToList();
        },
    });
    const { mutate: updateProjectVolume, isPending: isUpdatingProjectVolume } =
        ProjectClusterVolumesCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Volume updated");
                navigateToList();
            },
        });

    const isPending = isUpdatingClusterVolume || isUpdatingProjectVolume;

    function onSubmit(values: CreateOrEditVolumeFormOutput) {
        if (!canSubmit || !volume) {
            return;
        }

        if (scope.type === "cluster") {
            updateClusterVolume({
                volumeID: volume.id,
                payload: {
                    updateVer: volume.updateVer,
                    availableInProjects: values.availableInProjects,
                    default: values.default,
                },
            });
            return;
        }

        updateProjectVolume({
            projectID: scope.projectId,
            volumeID: volume.id,
            payload: {
                updateVer: volume.updateVer,
                default: values.default,
            },
        });
    }

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <RouteFormHeader title="Edit volume" />

            {isFetching ? (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            ) : volume ? (
                <CreateOrEditVolumeForm
                    key={volume.id}
                    initialValues={toVolumeFormInitialValues(volume)}
                    readOnlyCore
                    readOnlyAvailableInProjects={scope.type === "project" || !canSubmit}
                    readOnlyDefault={!canSubmit}
                    readOnlyInherited={isInherited}
                    readOnlyPermission={!canWrite}
                    isPending={isPending}
                    showAvailableInProjects={scope.type === "cluster"}
                    onSubmit={onSubmit}
                >
                    {!canSubmit ? (
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
                </CreateOrEditVolumeForm>
            ) : (
                <div className="py-10 text-center text-sm text-muted-foreground">Volume not found</div>
            )}
        </div>
    );
}

function getVolumeListRoute(scope: VolumeManagementScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.volumes.$route(scope.projectId);
    }

    return ROUTE.cluster.volumes.$route;
}

interface Props {
    scope: VolumeManagementScope;
    volumeId: string;
}
