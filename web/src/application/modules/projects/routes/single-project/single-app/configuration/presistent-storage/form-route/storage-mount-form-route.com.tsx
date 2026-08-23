import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import { AppStorageSettingsCommands, AppStorageSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { StorageMountForm } from "~/projects/dialogs/storage-mount/form";
import { formValuesToMount, mountToFormInput } from "~/projects/dialogs/storage-mount/form/storage-mount.form-mappers";
import type { StorageMountFormOutput } from "~/projects/dialogs/storage-mount/schemas";
import type { AppStorageMount } from "~/projects/domain";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type StorageMountWithId = AppStorageMount & { _id: string };

function buildMountsWithIds(mounts: AppStorageMount[]): StorageMountWithId[] {
    return mounts.map((mount, index) => ({
        ...mount,
        _id: mount.key ?? `mount-${index}`,
    }));
}

export function StorageMountFormRoute({ mode, projectId, env, appId, mountId }: Props) {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";

    const { data: appData, isLoading } = AppStorageSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutateAsync: update, isPending } = AppStorageSettingsCommands.useUpdateOne();

    const mountsWithIds = buildMountsWithIds(appData?.data.mounts ?? []);
    const mount = isEditMode ? mountsWithIds.find(item => item._id === mountId) : undefined;

    function navigateToList() {
        navigate.modules(
            ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(projectId, env, appId),
            {
                ignorePrevPath: true,
            },
        );
    }

    async function handleSubmit(values: StorageMountFormOutput) {
        if (!canWrite) {
            return;
        }

        const existingMounts = appData?.data.mounts ?? [];
        const updateVer = appData?.data.updateVer ?? 0;

        try {
            if (isEditMode) {
                if (!mountId) {
                    return;
                }

                const remainingMounts = mountsWithIds
                    .filter(item => item._id !== mountId)
                    .map(({ _id, ...item }) => item);

                await update({
                    projectID: projectId,
                    env,
                    appID: appId,
                    payload: {
                        mounts: [...remainingMounts, formValuesToMount(values)],
                        updateVer,
                    },
                });
                toast.success("Storage mount updated");
            } else {
                await update({
                    projectID: projectId,
                    env,
                    appID: appId,
                    payload: {
                        mounts: [...existingMounts, formValuesToMount(values)],
                        updateVer,
                    },
                });
                toast.success("Storage mount added");
            }

            navigateToList();
        } catch {
            toast.error(isEditMode ? "Failed to update storage mount" : "Failed to add storage mount");
        }
    }

    if (isLoading) {
        return <AppLoader />;
    }

    if (isEditMode && !mount) {
        return <div className="py-10 text-center text-sm text-muted-foreground">Storage mount not found</div>;
    }

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={isEditMode ? "Edit Storage" : "Add a new storage to the app"} />

            <StorageMountForm
                projectId={projectId}
                isPending={isPending}
                isEditMode={isEditMode}
                defaultValues={mount ? mountToFormInput(mount) : undefined}
                onSubmit={values => void handleSubmit(values)}
                readOnly={!canWrite}
                onClose={navigateToList}
            >
                <div className={cn(dashedBorderBox, "text-sm leading-6 mb-4")}>
                    <span className="font-semibold text-orange-500">Important:</span> If your cluster consists of more
                    than 1 node, you need to ensure that the directories or volumes are accessible from all nodes.
                    Otherwise, your apps may not function properly.
                </div>
            </StorageMountForm>
        </div>
    );
}

type Props = {
    mode: "create" | "edit";
    projectId: string;
    env: string;
    appId: string;
    mountId?: string;
};
