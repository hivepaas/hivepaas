import { useState } from "react";

import { toast } from "sonner";
import { AppConfigFilesCommands } from "~/projects/data/commands";
import { AppConfigFilesQueries } from "~/projects/data/queries";
import { CreateOrEditAppConfigFileForm } from "~/projects/dialogs/create-or-edit-app-config-file/form";
import type { CreateOrEditAppConfigFileFormOutput } from "~/projects/dialogs/create-or-edit-app-config-file/schemas";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type AppConfigFileFormRouteMode = "create" | "edit";

function fileToBase64(file: File): Promise<string> {
    return file.arrayBuffer().then(buffer => {
        let binary = "";
        const bytes = new Uint8Array(buffer);
        const chunkSize = 0x8000;

        for (let index = 0; index < bytes.length; index += chunkSize) {
            binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize));
        }

        return window.btoa(binary);
    });
}

async function getConfigFileContent(values: CreateOrEditAppConfigFileFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

function getSwarmRef(values: CreateOrEditAppConfigFileFormOutput) {
    if (!values.mountIntoFilesystem) {
        return undefined;
    }

    return {
        file: {
            name: values.filePath,
            mode: values.fileMode,
            uid: values.fileUid,
            gid: values.fileGid,
        },
    };
}

export function AppConfigFileFormRoute({ mode, projectId, appId, configFileId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.configFiles.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    const detailQuery = AppConfigFilesQueries.useFindOneById(
        {
            projectID: projectId,
            appID: appId,
            configFileID: configFileId ?? "",
        },
        {
            enabled: isEditMode && Boolean(configFileId),
        },
    );
    const configFile = detailQuery.data?.data;

    const { mutate: createAppConfigFile, isPending: isCreatingAppConfigFile } = AppConfigFilesCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App config file created successfully");
            navigateToList();
        },
    });

    const { mutate: updateAppConfigFile, isPending: isUpdatingAppConfigFile } = AppConfigFilesCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App config file updated successfully");
            navigateToList();
        },
    });

    async function onSubmit(values: CreateOrEditAppConfigFileFormOutput) {
        if (!canWrite) {
            return;
        }

        const content = await getConfigFileContent(values);
        const base64 = values.valueType === "binary";
        const swarmRef = getSwarmRef(values);

        if (isEditMode && configFile) {
            updateAppConfigFile({
                projectID: projectId,
                appID: appId,
                configFileID: configFile.id,
                updateVer: configFile.updateVer,
                name: values.name,
                content,
                base64,
                swarmRef,
            });
            return;
        }

        if (!isEditMode && content !== undefined) {
            createAppConfigFile({
                projectID: projectId,
                appID: appId,
                name: values.name,
                content,
                base64,
                swarmRef,
            });
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

    const isPending = isCreatingAppConfigFile || isUpdatingAppConfigFile;
    const initialValues =
        isEditMode && configFile
            ? {
                  name: configFile.name,
                  valueType: configFile.base64 ? ("binary" as const) : ("text" as const),
                  mountIntoFilesystem: Boolean(configFile.swarmRef?.file),
                  filePath: configFile.swarmRef?.file?.name,
                  fileMode: configFile.swarmRef?.file?.mode,
                  fileUid: configFile.swarmRef?.file?.uid,
                  fileGid: configFile.swarmRef?.file?.gid,
              }
            : undefined;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const shouldRenderForm = mode === "create" || Boolean(configFile);

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <RouteFormHeader title={mode === "create" ? "Create Config File" : "Edit Config File"} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditAppConfigFileForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    isEditMode={isEditMode}
                    initialValues={initialValues}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: AppConfigFileFormRouteMode;
    projectId: string;
    appId: string;
    configFileId?: string;
}
