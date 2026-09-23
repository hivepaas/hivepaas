import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import type { AppStorageFinding } from "~/projects/api/services";
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

import { StorageInUseDialog } from "../building-blocks";

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
    const { mutateAsync: preflight, isPending: isChecking } = AppStorageSettingsCommands.usePreflight();

    // Held while the operator answers for what is already in that directory.
    const [inUse, setInUse] = useState<{ mounts: AppStorageMount[]; findings: AppStorageFinding[] } | null>(null);

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

    async function save(mounts: AppStorageMount[], resetStorage: boolean) {
        try {
            await update({
                projectID: projectId,
                env,
                appID: appId,
                payload: {
                    mounts,
                    updateVer: appData?.data.updateVer ?? 0,
                    resetStorage: resetStorage || undefined,
                },
            });
            setInUse(null);
            toast.success(isEditMode ? "Storage mount updated" : "Storage mount added");
            navigateToList();
        } catch {
            // Error notification is handled by useAppStorageSettingsApi via notifyError
        }
    }

    async function handleSubmit(values: StorageMountFormOutput) {
        if (!canWrite) {
            return;
        }

        const existingMounts = appData?.data.mounts ?? [];
        const keptMounts = isEditMode
            ? mountsWithIds.filter(item => item._id !== mountId).map(({ _id, ...item }) => item)
            : existingMounts;
        if (isEditMode && !mountId) {
            return;
        }
        const mounts = [...keptMounts, formValuesToMount(values)];

        // What is already in the directory this mount would reach. The check is
        // advisory: one that cannot answer must not stop a save that would have
        // worked.
        try {
            const result = await preflight({ projectID: projectId, env, appID: appId, payload: { mounts } });
            // Storage nothing could be seen of is said out loud rather than read
            // as "there is nothing there".
            if (result.data.unchecked.length > 0) {
                toast.warning("Could not check what is already in that storage.", {
                    description: "If this app ran here before, its data is still in place.",
                });
            }
            if (result.data.storage.length > 0) {
                setInUse({ mounts, findings: result.data.storage });
                return;
            }
        } catch {
            // Fall through to saving, which reports its own failures.
        }

        await save(mounts, false);
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
                env={env}
                appId={appId}
                isPending={isPending || isChecking}
                isEditMode={isEditMode}
                defaultValues={mount ? mountToFormInput(mount) : undefined}
                onSubmit={values => void handleSubmit(values)}
                readOnly={!canWrite}
                onClose={navigateToList}
            >
                <div className={cn(dashedBorderBox, "mb-4")}>
                    <span className="font-semibold text-orange-500">Important:</span> If your cluster consists of more
                    than 1 node, you need to ensure that the directories or volumes are accessible from all nodes.
                    Otherwise, your apps may not function properly.
                </div>
            </StorageMountForm>

            <StorageInUseDialog
                open={inUse !== null}
                findings={inUse?.findings ?? []}
                isPending={isPending}
                onOpenChange={nextOpen => {
                    if (!nextOpen) {
                        setInUse(null);
                    }
                }}
                onConfirm={resetStorage => {
                    if (inUse) {
                        void save(inUse.mounts, resetStorage);
                    }
                }}
            />
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
