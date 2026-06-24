import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppStorageSettingsCommands, AppStorageSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { StorageMountForm } from "~/projects/dialogs/storage-mount/form";
import { formValuesToMount } from "~/projects/dialogs/storage-mount/form/storage-mount.form-mappers";
import type { StorageMountFormOutput } from "~/projects/dialogs/storage-mount/schemas";
import type { AppStorageMount } from "~/projects/domain";

import { RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { useStorageMounts } from "../context";

type StorageMountWithId = AppStorageMount & { _id: string };
type StorageMountFormRouteMode = "create" | "edit";

function mountWithoutId(mount: StorageMountWithId): AppStorageMount {
    const { _id: _unused, ...rest } = mount;
    return rest;
}

export function StorageMountFormRoute({ mode, projectId, appId, mountId }: Props) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const { mounts, addMount, updateMount } = useStorageMounts();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";
    const targetMount = isEditMode && mountId ? mounts.find(mount => mount._id === mountId) : undefined;

    const { data: appData } = AppStorageSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );
    const { mutateAsync: update } = AppStorageSettingsCommands.useUpdateOne();
    const updateVer = appData?.data.updateVer ?? 0;

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    async function persistMounts(nextMounts: StorageMountWithId[], successMessage: string) {
        if (!canWrite) {
            return;
        }

        const mountsWithoutIds = nextMounts.map(({ _id, ...mount }) => mount);

        await update({
            projectID: projectId,
            appID: appId,
            payload: {
                mounts: mountsWithoutIds,
                updateVer,
            },
        });
        toast.success(successMessage);
    }

    async function handleSubmit(values: StorageMountFormOutput) {
        if (!canWrite) {
            return;
        }

        const nextMount = formValuesToMount(values);

        try {
            setIsSubmitting(true);

            if (isEditMode) {
                invariant(targetMount, "targetMount must be defined");

                const nextMounts = mounts.map(existing =>
                    existing._id === targetMount._id ? { ...nextMount, _id: targetMount._id } : existing,
                );
                await persistMounts(nextMounts, "Storage mount updated");
                updateMount(targetMount._id, nextMount);
                navigateToList();
                return;
            }

            const mountWithId: StorageMountWithId = {
                ...nextMount,
                _id: `mount-${Date.now()}-${Math.random()}`,
            };
            await persistMounts([...mounts, mountWithId], "Storage mount created");
            addMount(nextMount);
            navigateToList();
        } finally {
            setIsSubmitting(false);
        }
    }

    const shouldRenderForm = !isEditMode || Boolean(targetMount);

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <RouteFormHeader title={isEditMode ? "Edit Storage" : "Add a new storage to the app"} />

            {shouldRenderForm ? (
                <StorageMountForm
                    isPending={isSubmitting}
                    onSubmit={values => void handleSubmit(values)}
                    defaultValues={targetMount ? mountWithoutId(targetMount) : undefined}
                    readOnly={!canWrite}
                    onClose={navigateToList}
                >
                    <div className={cn(dashedBorderBox, "text-[12px] text-center")}>
                        <span className="font-bold text-orange-500">Important:</span> If your cluster consists of more
                        than 1 node, you need to ensure that the directories or volumes are accessible from all nodes.
                        Otherwise, your apps may not function properly.
                    </div>
                </StorageMountForm>
            ) : (
                <div className="py-10 text-center text-sm text-muted-foreground">Storage mount not found</div>
            )}
        </div>
    );
}

interface Props {
    mode: StorageMountFormRouteMode;
    projectId: string;
    appId: string;
    mountId?: string;
}
