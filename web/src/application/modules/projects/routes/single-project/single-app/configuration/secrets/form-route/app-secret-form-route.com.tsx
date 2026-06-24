import { useState } from "react";

import { toast } from "sonner";
import { ProjectAppSecretsCommands } from "~/projects/data/commands";
import { ProjectAppSecretsQueries } from "~/projects/data/queries";
import { CreateOrEditAppSecretForm } from "~/projects/dialogs/create-or-edit-app-secret/form";
import type { CreateOrEditAppSecretFormOutput } from "~/projects/dialogs/create-or-edit-app-secret/schemas";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type AppSecretFormRouteMode = "create" | "edit";

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

async function getSecretValue(values: CreateOrEditAppSecretFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

function getSwarmRef(values: CreateOrEditAppSecretFormOutput) {
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

export function AppSecretFormRoute({ mode, projectId, appId, secretId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    const detailQuery = ProjectAppSecretsQueries.useFindOneById(
        {
            projectID: projectId,
            appID: appId,
            secretID: secretId ?? "",
        },
        {
            enabled: isEditMode && Boolean(secretId),
        },
    );
    const secret = detailQuery.data?.data;

    const { mutate: createAppSecret, isPending: isCreatingApp } = ProjectAppSecretsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App secret created successfully");
            navigateToList();
        },
    });

    const { mutate: updateAppSecret, isPending: isUpdatingApp } = ProjectAppSecretsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App secret updated successfully");
            navigateToList();
        },
    });

    async function onSubmit(values: CreateOrEditAppSecretFormOutput) {
        if (!canWrite) {
            return;
        }

        const value = await getSecretValue(values);
        const base64 = values.valueType === "binary";
        const swarmRef = getSwarmRef(values);

        if (isEditMode && secret) {
            updateAppSecret({
                projectID: projectId,
                appID: appId,
                secretID: secret.id,
                updateVer: secret.updateVer,
                name: values.name,
                value,
                base64,
                swarmRef,
            });
            return;
        }

        if (!isEditMode && value !== undefined) {
            createAppSecret({
                projectID: projectId,
                appID: appId,
                name: values.name,
                value,
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

    const isPending = isCreatingApp || isUpdatingApp;
    const initialValues =
        isEditMode && secret
            ? {
                  name: secret.name,
                  valueType: secret.base64 ? ("binary" as const) : ("text" as const),
                  mountIntoFilesystem: Boolean(secret.swarmRef?.file),
                  filePath: secret.swarmRef?.file?.name,
                  fileMode: secret.swarmRef?.file?.mode,
                  fileUid: secret.swarmRef?.file?.uid,
                  fileGid: secret.swarmRef?.file?.gid,
              }
            : undefined;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const shouldRenderForm = mode === "create" || Boolean(secret);

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <RouteFormHeader title={mode === "create" ? "Create Secret" : "Edit Secret"} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditAppSecretForm
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
    mode: AppSecretFormRouteMode;
    projectId: string;
    appId: string;
    secretId?: string;
}
