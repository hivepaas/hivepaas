import { Button } from "@components/ui";
import { toast } from "sonner";
import { ClusterVolumesCommands } from "~/cluster/data/commands";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";
import { ProjectClusterVolumesCommands } from "~/projects/data/commands";

import { FormActionBar, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { CreateOrEditVolumeForm, type CreateOrEditVolumeFormOutput } from "../volume-form";

import { toVolumeBasePayload, toVolumeCreatePayload } from "./volume-form-route.helpers";

export function CreateVolumeFormRoute({ scope }: Props) {
    const { navigate } = useAppNavigate();
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const canWrite = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;

    function navigateToList() {
        navigate.modules(getVolumeListRoute(scope), { ignorePrevPath: true });
    }

    const { mutate: createClusterVolume, isPending: isCreatingClusterVolume } = ClusterVolumesCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Volume created");
            navigateToList();
        },
    });
    const { mutate: createProjectVolume, isPending: isCreatingProjectVolume } =
        ProjectClusterVolumesCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Volume created");
                navigateToList();
            },
        });

    const isPending = isCreatingClusterVolume || isCreatingProjectVolume;

    function onSubmit(values: CreateOrEditVolumeFormOutput) {
        if (!canWrite) {
            return;
        }

        if (scope.type === "cluster") {
            createClusterVolume({
                payload: toVolumeCreatePayload(values, values.availableInProjects),
            });
            return;
        }

        createProjectVolume({
            projectID: scope.projectId,
            payload: {
                ...toVolumeBasePayload(values),
                availableInProjects: false,
                default: values.default,
            },
        });
    }

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title="Create a volume" />

            <CreateOrEditVolumeForm
                readOnlyCore={!canWrite}
                readOnlyAvailableInProjects={!canWrite}
                readOnlyDefault={!canWrite}
                readOnlyPermission={!canWrite}
                isPending={isPending}
                showAvailableInProjects={scope.type === "cluster"}
                onSubmit={onSubmit}
            >
                {!canWrite ? (
                    <FormActionBar>
                        <Button
                            type="button"
                            onClick={navigateToList}
                        >
                            Close
                        </Button>
                    </FormActionBar>
                ) : (
                    <FormActionBar>
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
                    </FormActionBar>
                )}
            </CreateOrEditVolumeForm>
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
}
