import { useState } from "react";

import { toast } from "sonner";
import type { ProjectCommandPipe_CreateOne_Payload } from "~/projects/api/services";
import { ProjectCommandPipeCommands } from "~/projects/data/commands";
import { ProjectCommandPipeQueries } from "~/projects/data/queries";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { ProjectCommandPipeForm } from "../form";
import type { ProjectCommandPipeFormOutput } from "../schemas";

type ProjectCommandPipeFormRouteMode = "create" | "edit";

export function ProjectCommandPipeFormRoute({ mode, projectId, env, commandPipeId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions({ type: "project" });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (commandPipeId ?? "") : "";
    const listRoute = ROUTE.projects.single.providerConfiguration.commandPipes.$route(projectId);

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createProjectCommandPipe, isPending: isCreating } = ProjectCommandPipeCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project Command Pipe created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectCommandPipe, isPending: isUpdating } = ProjectCommandPipeCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project Command Pipe updated successfully");
            markSaved();
        },
    });

    const detailQuery = ProjectCommandPipeQueries.useFindOneById(
        {
            projectID: projectId,
            env,
            id: detailId,
        },
        { enabled: isEditMode },
    );
    const commandPipe = detailQuery.data?.data;
    const readOnlyInherited = commandPipe?.inherited === true;

    function createPayload(values: ProjectCommandPipeFormOutput): ProjectCommandPipe_CreateOne_Payload {
        return {
            inheritable: false,
            default: values.default,
            name: values.name,
            sourceCommand: { id: values.sourceCommandId },
            targetCommand: { id: values.targetCommandId },
        };
    }

    function onSubmit(values: ProjectCommandPipeFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && commandPipe) {
            updateProjectCommandPipe({
                projectID: projectId,
                env,
                id: commandPipe.id,
                payload: {
                    ...payload,
                    updateVer: commandPipe.updateVer,
                },
            });
            return;
        }

        createProjectCommandPipe({ projectID: projectId, env, payload });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        ) {
            return;
        }

        navigateToList();
    }

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const shouldRenderForm = mode === "create" || commandPipe;
    const title = mode === "create" ? "Create Command Pipe" : "Edit Command Pipe";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <ProjectCommandPipeForm
                    projectId={projectId}
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={commandPipe}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: ProjectCommandPipeFormRouteMode;
    projectId: string;
    env?: string;
    commandPipeId?: string;
}
