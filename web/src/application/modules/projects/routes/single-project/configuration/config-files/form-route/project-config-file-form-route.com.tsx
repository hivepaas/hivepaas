import { useState } from "react";

import { toast } from "sonner";
import { ProjectConfigFilesCommands } from "~/projects/data/commands";
import { ProjectConfigFilesQueries } from "~/projects/data/queries";
import { CreateOrEditAppConfigFileForm } from "~/projects/dialogs/create-or-edit-app-config-file/form";
import type { CreateOrEditAppConfigFileFormOutput } from "~/projects/dialogs/create-or-edit-app-config-file/schemas";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type ProjectConfigFileFormRouteMode = "create" | "edit";

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

async function getConfigFileContent(values: CreateOrEditAppConfigFileFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

export function ProjectConfigFileFormRoute({ mode, projectId, configFileId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.providerConfiguration.configFiles.$route(projectId), {
            ignorePrevPath: true,
        });
    }

    const detailQuery = ProjectConfigFilesQueries.useFindOneById(
        {
            projectID: projectId,
            env: scopedEnv,
            configFileID: configFileId ?? "",
        },
        {
            enabled: isEditMode && Boolean(configFileId),
        },
    );
    const configFile = detailQuery.data?.data;
    // An env shows its project's config files too; they are changed at the project.
    const isInherited = Boolean(configFile?.inherited);

    const { mutate: createOne, isPending: isCreating } = ProjectConfigFilesCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project config file created successfully");
            navigateToList();
        },
    });

    const { mutate: updateOne, isPending: isUpdating } = ProjectConfigFilesCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project config file updated successfully");
            navigateToList();
        },
    });

    async function onSubmit(values: CreateOrEditAppConfigFileFormOutput) {
        if (!canWrite || isInherited) {
            return;
        }

        const content = await getConfigFileContent(values);
        const base64 = values.valueType === "binary";

        if (isEditMode && configFile) {
            updateOne({
                projectID: projectId,
                env: scopedEnv,
                configFileID: configFile.id,
                updateVer: configFile.updateVer,
                name: values.name,
                content,
                base64,
                inheritable: values.inheritable,
            });
            return;
        }

        if (!isEditMode && content !== undefined) {
            createOne({
                projectID: projectId,
                env: scopedEnv,
                name: values.name,
                content,
                base64,
                inheritable: values.inheritable,
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

    const initialValues =
        isEditMode && configFile
            ? {
                  name: configFile.name,
                  valueType: configFile.base64 ? ("binary" as const) : ("text" as const),
                  inheritable: configFile.inheritable,
              }
            : undefined;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const shouldRenderForm = mode === "create" || Boolean(configFile);

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={mode === "create" ? "Create Config File" : "Edit Config File"} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditAppConfigFileForm
                    isPending={isCreating || isUpdating}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    isEditMode={isEditMode}
                    initialValues={initialValues}
                    readOnly={!canWrite || isInherited}
                    stickyActions
                    onClose={handleClose}
                    inheritableLabel="Available in Apps"
                    inheritableWarning="Warning: Apps will not be able to access this configuration."
                />
            )}
        </div>
    );
}

interface Props {
    mode: ProjectConfigFileFormRouteMode;
    projectId: string;
    configFileId?: string;
}
