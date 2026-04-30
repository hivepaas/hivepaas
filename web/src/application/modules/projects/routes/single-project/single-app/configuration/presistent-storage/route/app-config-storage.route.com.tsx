import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppStorageSettingsCommands, AppStorageSettingsQueries, ProjectStorageSettingsQueries } from "~/projects/data";
import { useStorageMountDialog } from "~/projects/dialogs/storage-mount";
import type { AppStorageMount } from "~/projects/domain";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { StorageTable } from "../building-blocks";
import { StorageMountsProvider, useStorageMounts } from "../context";

type StorageMountWithId = AppStorageMount & { _id: string };

function AppConfigStorageContent() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const { mounts, addMount, updateMount } = useStorageMounts();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data: appData, isLoading: appLoading } = AppStorageSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { data: projectRulesData, isLoading: projectRulesLoading } = ProjectStorageSettingsQueries.useFindOne({
        projectID: projectId,
    });

    const { mutate: update, isPending } = AppStorageSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Storage settings updated");
        },
        onError: err => {
            if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update storage settings");
            }
        },
    });

    const storageMountDialog = useStorageMountDialog();

    const handleAddMount = () => {
        storageMountDialog.actions.open(projectRulesData?.data, {
            onSubmit: (mount: AppStorageMount) => {
                addMount(mount);
            },
        });
    };

    const handleEditMount = (mount: StorageMountWithId) => {
        storageMountDialog.actions.openEdit(mount, projectRulesData?.data, {
            onSubmit: (updatedMount: AppStorageMount) => {
                updateMount(mount._id, updatedMount);
            },
        });
    };

    const handleSave = () => {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        const mountsWithoutIds = mounts.map(({ _id, ...mount }) => mount);

        update({
            projectID: projectId,
            appID: appId,
            payload: {
                mounts: mountsWithoutIds,
                updateVer: appData?.data.updateVer ?? 0,
            },
        });
    };

    if (appLoading || projectRulesLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="rounded-md border border-dashed border-primary bg-accent p-2 text-sm text-muted-foreground">
                For configuration details, see{" "}
                <a
                    href="https://docs.docker.com/reference/api/engine/version/v1.52/#tag/Service/operation/ServiceUpdate"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-500 underline underline-offset-2"
                >
                    docs
                </a>
            </div>

            <div className="rounded-lg border p-4">
                <StorageTable
                    onAddMount={handleAddMount}
                    onEditMount={handleEditMount}
                />
            </div>

            <div className="flex justify-end mt-4">
                <Button
                    onClick={handleSave}
                    className="min-w-[100px]"
                    disabled={isPending}
                    isLoading={isPending}
                >
                    Save
                </Button>
            </div>
        </div>
    );
}

export function AppConfigStorageRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = AppStorageSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={refetch}
            />
        );
    }

    return (
        <StorageMountsProvider initialMounts={data?.data.mounts ?? []}>
            <AppConfigStorageContent />
        </StorageMountsProvider>
    );
}
