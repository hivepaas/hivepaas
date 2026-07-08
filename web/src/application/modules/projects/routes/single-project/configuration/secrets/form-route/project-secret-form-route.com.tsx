import { useState } from "react";

import { toast } from "sonner";
import { ProjectSecretsCommands } from "~/projects/data/commands";
import { ProjectSecretsQueries } from "~/projects/data/queries";
import { CreateOrEditProjectSecretForm } from "~/projects/dialogs/create-or-edit-project-secret/form";
import type { CreateOrEditProjectSecretFormOutput } from "~/projects/dialogs/create-or-edit-project-secret/schemas";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type ProjectSecretFormRouteMode = "create" | "edit";

async function fileToBase64(file: File): Promise<string> {
    const buffer = await file.arrayBuffer();
    const bytes = new Uint8Array(buffer);
    const chunkSize = 0x8000;
    let binary = "";

    for (let index = 0; index < bytes.length; index += chunkSize) {
        binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize));
    }

    return window.btoa(binary);
}

async function getSecretValue(values: CreateOrEditProjectSecretFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

export function ProjectSecretFormRoute({ mode, projectId, secretId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";
    const listRoute = ROUTE.projects.single.providerConfiguration.secrets.$route(projectId);

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const detailQuery = ProjectSecretsQueries.useFindOneById(
        {
            projectID: projectId,
            secretID: secretId ?? "",
        },
        {
            enabled: isEditMode && Boolean(secretId),
        },
    );
    const secret = detailQuery.data?.data;

    const { mutate: createProjectSecret, isPending: isCreatingProject } = ProjectSecretsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project secret created successfully");
            markSaved();
        },
    });

    const { mutate: updateProjectSecret, isPending: isUpdatingProject } = ProjectSecretsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project secret updated successfully");
            markSaved();
        },
    });

    async function onSubmit(values: CreateOrEditProjectSecretFormOutput) {
        if (!canWrite) {
            return;
        }

        const value = await getSecretValue(values);
        const base64 = values.valueType === "binary";

        if (isEditMode && secret) {
            updateProjectSecret({
                projectID: projectId,
                secretID: secret.id,
                updateVer: secret.updateVer,
                name: values.name,
                value,
                base64,
            });
            return;
        }

        if (!isEditMode && value !== undefined) {
            createProjectSecret({
                projectID: projectId,
                name: values.name,
                value,
                base64,
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

    const isPending = isCreatingProject || isUpdatingProject;
    const initialValues =
        isEditMode && secret
            ? {
                  name: secret.name,
                  valueType: secret.base64 ? ("binary" as const) : ("text" as const),
              }
            : undefined;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const shouldRenderForm = mode === "create" || Boolean(secret);

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={mode === "create" ? "Create Secret" : "Edit Secret"} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditProjectSecretForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    isEditMode={isEditMode}
                    initialValues={initialValues}
                    readOnly={!canWrite}
                    stickyActions
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: ProjectSecretFormRouteMode;
    projectId: string;
    secretId?: string;
}
