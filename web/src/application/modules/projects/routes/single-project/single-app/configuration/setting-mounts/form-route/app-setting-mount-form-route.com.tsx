import { useState } from "react";

import { toast } from "sonner";
import { AppSettingMountsCommands } from "~/projects/data/commands";
import { AppSettingMountsQueries } from "~/projects/data/queries";
import type { AppSettingMount, AppSettingMountSource } from "~/projects/domain";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { AppSettingMountForm } from "../form";
import type { AppSettingMountFormInput, AppSettingMountFormOutput } from "../schemas";
import { type Grant, rowsFor } from "../utils";

type AppSettingMountFormRouteMode = "create" | "edit";

/** The entry as the form edits it: a row for every part of its type, ticked where it has the part. */
function toFormValues(sources: AppSettingMountSource[], entry: AppSettingMount): AppSettingMountFormInput {
    const { type } = entry.source;
    return {
        name: entry.name,
        inheritable: entry.inheritable,
        sourceType: type,
        source: { id: entry.source.id, name: entry.source.name },
        files: rowsFor(sources, type, entry.source.name).map(row => {
            const file = entry.files.find(item => item.part === row.part);
            return file
                ? { ...row, enabled: true, path: file.path, mode: file.mode, uid: file.uid, gid: file.gid }
                : row;
        }),
    };
}

/** What the entry hands out as loaded: nothing when it is disabled. */
function grantsOfEntry(entry: AppSettingMount | undefined): Grant[] {
    if (!entry || entry.status !== EProjectSecretStatus.Active) {
        return [];
    }
    return entry.files.filter(file => file.gated).map(file => ({ source: entry.source.id, part: file.part }));
}

function toPayload(values: AppSettingMountFormOutput) {
    return {
        name: values.name,
        inheritable: values.inheritable,
        sourceID: values.source?.id ?? "",
        files: values.files
            .filter(file => file.enabled)
            .map(file => ({
                part: file.part,
                path: file.path.trim(),
                uid: file.uid.trim(),
                gid: file.gid.trim(),
                mode: file.mode.trim(),
            })),
    };
}

export function AppSettingMountFormRoute({ mode, projectId, appId, env, settingMountId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.settingMounts.$route(projectId, env, appId), {
            ignorePrevPath: true,
        });
    }

    const sourcesQuery = AppSettingMountsQueries.useFindSources({ projectID: projectId, env, appID: appId });
    const detailQuery = AppSettingMountsQueries.useFindOneById(
        { projectID: projectId, env, appID: appId, settingMountID: settingMountId ?? "" },
        { enabled: isEditMode && Boolean(settingMountId) },
    );
    const sources = sourcesQuery.data?.data.sources;
    const mayMountSensitive = sourcesQuery.data?.data.mayMountSensitive ?? false;
    const entry = detailQuery.data?.data;

    const { mutate: createOne, isPending: isCreating } = AppSettingMountsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Setting mount created");
            navigateToList();
        },
    });
    const { mutate: updateOne, isPending: isUpdating } = AppSettingMountsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Setting mount updated");
            navigateToList();
        },
    });

    function onSubmit(values: AppSettingMountFormOutput) {
        if (!canWrite) {
            return;
        }
        if (isEditMode && entry) {
            updateOne({
                projectID: projectId,
                env,
                appID: appId,
                settingMountID: entry.id,
                updateVer: entry.updateVer,
                ...toPayload(values),
            });
            return;
        }
        if (!isEditMode) {
            createOne({ projectID: projectId, env, appID: appId, ...toPayload(values) });
        }
    }

    function handleClose(): void {
        if (canWrite && hasChanges) {
            const userConfirmed = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }
        setHasChanges(false);
        navigateToList();
    }

    const isLoading = sourcesQuery.isFetching || (isEditMode && detailQuery.isFetching);
    const shouldRenderForm = Boolean(sources) && (!isEditMode || Boolean(entry));

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={isEditMode ? "Edit Setting Mount" : "Create Setting Mount"} />

            {isLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isLoading && shouldRenderForm && sources && (
                <AppSettingMountForm
                    sources={sources}
                    mayMountSensitive={mayMountSensitive}
                    isPending={isCreating || isUpdating}
                    initialValues={entry ? toFormValues(sources, entry) : undefined}
                    initialGrants={grantsOfEntry(entry)}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    readOnly={!canWrite}
                    stickyActions
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: AppSettingMountFormRouteMode;
    projectId: string;
    env: string;
    appId: string;
    settingMountId?: string;
}
