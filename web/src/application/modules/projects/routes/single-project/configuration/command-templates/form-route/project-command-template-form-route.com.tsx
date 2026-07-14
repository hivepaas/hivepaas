import { useState } from "react";

import { toast } from "sonner";
import type { ProjectCommandTemplate_CreateOne_Payload } from "~/projects/api/services";
import { ProjectCommandTemplateCommands } from "~/projects/data/commands";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import {
    ProjectCommandTemplateForm,
    normalizeProjectCommandTemplateArgGroups,
    normalizeProjectCommandTemplateEnvVars,
} from "../form";
import { PROJECT_COMMAND_TEMPLATE_COMMAND_MODE, type ProjectCommandTemplateFormOutput } from "../schemas";

type ProjectCommandTemplateFormRouteMode = "create" | "edit";

export function ProjectCommandTemplateFormRoute({ mode, projectId, commandTemplateId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions({ type: "project" });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (commandTemplateId ?? "") : "";
    const listRoute = ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId);

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createProjectCommandTemplate, isPending: isCreating } = ProjectCommandTemplateCommands.useCreateOne(
        {
            onSuccess: () => {
                toast.success("Project Command Template created successfully");
                markSaved();
            },
        },
    );
    const { mutate: updateProjectCommandTemplate, isPending: isUpdating } = ProjectCommandTemplateCommands.useUpdateOne(
        {
            onSuccess: () => {
                toast.success("Project Command Template updated successfully");
                markSaved();
            },
        },
    );

    const detailQuery = ProjectCommandTemplateQueries.useFindOneById(
        {
            projectID: projectId,
            id: detailId,
        },
        { enabled: isEditMode },
    );
    const commandTemplate = detailQuery.data?.data;
    const readOnlyInherited = commandTemplate?.inherited === true;

    function createPayload(values: ProjectCommandTemplateFormOutput): ProjectCommandTemplate_CreateOne_Payload {
        return {
            availableInProjects: false,
            default: values.default,
            name: values.name,
            kind: values.kind,
            command: values.commandMode === PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Command ? values.command : "",
            script: values.commandMode === PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script ? values.script : "",
            workingDir: values.workingDir,
            envVars: normalizeProjectCommandTemplateEnvVars(values.envVars),
            argGroups: normalizeProjectCommandTemplateArgGroups(values.argGroups),
            consoleSize: values.consoleSize,
            tty: values.tty,
        };
    }

    function onSubmit(values: ProjectCommandTemplateFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && commandTemplate) {
            updateProjectCommandTemplate({
                projectID: projectId,
                id: commandTemplate.id,
                payload: {
                    ...payload,
                    updateVer: commandTemplate.updateVer,
                },
            });
            return;
        }

        createProjectCommandTemplate({ projectID: projectId, payload });
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
    const shouldRenderForm = mode === "create" || commandTemplate;
    const title = mode === "create" ? "Create Command Template" : "Edit Command Template";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <ProjectCommandTemplateForm
                    projectId={projectId}
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={commandTemplate}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: ProjectCommandTemplateFormRouteMode;
    projectId: string;
    commandTemplateId?: string;
}
