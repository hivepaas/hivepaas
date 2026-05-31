import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectRepoWebhookCommands } from "~/projects/data/commands";
import { ProjectRepoWebhookQueries } from "~/projects/data/queries";
import { RepoWebhookCommands } from "~/settings/data/commands";
import { RepoWebhookQueries } from "~/settings/data/queries";
import type { ERepoWebhookKind } from "~/settings/module-shared/enums";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";

import { CreateOrEditRepoWebhookForm } from "../form";
import { useCreateOrEditRepoWebhookDialogState } from "../hooks";
import type { CreateOrEditRepoWebhookFormOutput } from "../schemas";

type CreatedWebhookResult = {
    id: string;
    secret: string;
    webhookURL: string;
};

export function CreateOrEditRepoWebhookDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
        openEdit,
    } = useCreateOrEditRepoWebhookDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [createdWebhook, setCreatedWebhook] = useState<CreatedWebhookResult | null>(null);

    const permissionScope = state.mode === "closed" ? ({ type: "settings" } as const) : state.scope;
    const { canWrite } = useSettingsScopePermissions(permissionScope);

    const { mutate: createSettingsRepoWebhook, isPending: isCreatingSettings } = RepoWebhookCommands.useCreateOne({
        onSuccess: (response, variables) => {
            toast.success("Webhook created successfully");
            setCreatedWebhook(response.data);
            openEdit({ type: "settings" }, response.data.id, { props: dialogOptions });
            dialogOptions?.onSuccess?.();
            void variables;
        },
    });
    const { mutate: updateSettingsRepoWebhook, isPending: isUpdatingSettings } = RepoWebhookCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Webhook updated successfully");
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectRepoWebhook, isPending: isCreatingProject } = ProjectRepoWebhookCommands.useCreateOne({
        onSuccess: (response, variables) => {
            toast.success("Project webhook created successfully");
            setCreatedWebhook(response.data);
            openEdit({ type: "project", projectId: variables.projectID }, response.data.id, { props: dialogOptions });
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectRepoWebhook, isPending: isUpdatingProject } = ProjectRepoWebhookCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project webhook updated successfully");
            dialogOptions?.onSuccess?.();
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setCreatedWebhook(null);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "edit" ? state.id : "";
    const settingsDetailQuery = RepoWebhookQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectRepoWebhookQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "edit" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingsDetailQuery;
    const repoWebhook = detailQuery.data?.data;

    function createPayload(values: CreateOrEditRepoWebhookFormOutput) {
        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            kind: values.kind,
            secret: values.secret,
        };
    }

    function onSubmit(values: CreateOrEditRepoWebhookFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values);

        if (state.mode === "edit" && repoWebhook) {
            const updatePayload = {
                ...payload,
                updateVer: repoWebhook.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectRepoWebhook({
                    projectID: state.scope.projectId,
                    id: repoWebhook.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingsRepoWebhook({
                id: repoWebhook.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectRepoWebhook({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingsRepoWebhook({ payload });
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

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited = resolvedDialogOptions.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited
        ? (resolvedDialogOptions.entityTitle ?? "Webhook")
        : "Create or update a webhook";
    const isPending = isCreatingSettings || isUpdatingSettings || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const matchedCreatedWebhook =
        createdWebhook && state.mode === "edit" && createdWebhook.id === state.id ? createdWebhook : null;
    const initialValues = repoWebhook
        ? {
              name: repoWebhook.name,
              kind: repoWebhook.kind as ERepoWebhookKind,
              secret: matchedCreatedWebhook?.secret ?? repoWebhook.secret,
              availableInProjects: repoWebhook.availableInProjects ?? false,
              default: repoWebhook.default ?? false,
          }
        : undefined;
    const webhookURL = repoWebhook?.webhookURL ?? matchedCreatedWebhook?.webhookURL;
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;
    const canRenderForm = state.mode === "open" || (state.mode === "edit" && !!repoWebhook);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && canRenderForm && (
                    <CreateOrEditRepoWebhookForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        webhookURL={webhookURL}
                        showAvailableInProjects={showAvailableInProjects}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
