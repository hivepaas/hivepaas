import { Outlet, useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppStorageSettingsCommands, AppStorageSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import type { AppStorageMount } from "~/projects/domain";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { StorageTable } from "../building-blocks";
import { StorageMountsProvider, useStorageMounts } from "../context";

type StorageMountWithId = AppStorageMount & { _id: string };

function AppConfigStorageContent() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const { mounts, removeMount } = useStorageMounts();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data: appData, isLoading: appLoading } = AppStorageSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutateAsync: update } = AppStorageSettingsCommands.useUpdateOne();

    const updateVer = appData?.data.updateVer ?? 0;

    async function persistMounts(nextMounts: StorageMountWithId[], successMessage: string) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        const mountsWithoutIds = nextMounts.map(({ _id, ...mount }) => mount);

        try {
            await update({
                projectID: projectId,
                appID: appId,
                payload: {
                    mounts: mountsWithoutIds,
                    updateVer,
                },
            });
            toast.success(successMessage);
        } catch {
            // toast.error("Failed to update storage settings");
            throw new Error("Failed to update storage settings");
        }
    }

    const handleAddMount = () => {
        if (!canWrite) {
            return;
        }

        navigate.modules(
            ROUTE.projects.single.apps.single.configuration.presistentStorage.create.$route(projectId, appId),
        );
    };

    const handleEditMount = (mount: StorageMountWithId) => {
        navigate.modules(
            ROUTE.projects.single.apps.single.configuration.presistentStorage.edit.$route(projectId, appId, mount._id),
        );
    };

    const handleDeleteMount = (mount: StorageMountWithId) => {
        if (!canWrite) {
            return;
        }

        const nextMounts = mounts.filter(existing => existing._id !== mount._id);
        void persistMounts(nextMounts, "Storage mount removed").then(() => {
            removeMount(mount._id);
        });
    };

    if (appLoading) {
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

            <StorageTable
                onAddMount={handleAddMount}
                onEditMount={handleEditMount}
                onDeleteMount={handleDeleteMount}
                canWrite={canWrite}
            />
        </div>
    );
}

export function AppConfigStorageListRoute() {
    return <AppConfigStorageContent />;
}

export function AppConfigStorageRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = AppStorageSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

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
            <Outlet />
        </StorageMountsProvider>
    );
}
