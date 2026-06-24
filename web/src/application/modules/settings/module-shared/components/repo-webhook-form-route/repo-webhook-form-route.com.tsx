import { useState } from "react";

import { useLocation } from "react-router";
import { toast } from "sonner";
import { ProjectRepoWebhookCommands } from "~/projects/data/commands";
import { ProjectRepoWebhookQueries } from "~/projects/data/queries";
import { RepoWebhookCommands } from "~/settings/data/commands";
import { RepoWebhookQueries } from "~/settings/data/queries";
import { CreateOrEditRepoWebhookForm } from "~/settings/dialogs/create-or-edit-repo-webhook/form";
import type { CreateOrEditRepoWebhookFormOutput } from "~/settings/dialogs/create-or-edit-repo-webhook/schemas";
import type { ERepoWebhookKind } from "~/settings/module-shared/enums";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { RepoWebhookTableScope } from "../repo-webhook-table";
import { SettingsFormRouteHeader } from "../settings-form-route-header";

type RepoWebhookFormRouteMode = "create" | "edit";

type CreatedWebhookResult = {
    id: string;
    secret: string;
    webhookURL: string;
};

type RepoWebhookLocationState = {
    createdRepoWebhook?: CreatedWebhookResult;
};

export function RepoWebhookFormRoute({ mode, scope, repoWebhookId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();
    const location = useLocation();

    const listRoute = getRepoWebhookListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (repoWebhookId ?? "") : "";
    const createdWebhook = (location.state as RepoWebhookLocationState | null)?.createdRepoWebhook ?? null;

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function navigateToEdit(created: CreatedWebhookResult) {
        const editRoute =
            scope.type === "project"
                ? ROUTE.projects.single.providerConfiguration.webhooks.edit.$route(scope.projectId, created.id)
                : ROUTE.sources.webhooks.edit.$route(created.id);

        navigate.modules(editRoute, {
            ignorePrevPath: true,
            state: {
                createdRepoWebhook: created,
            },
        });
    }

    const { mutate: createSettingsRepoWebhook, isPending: isCreatingSettings } = RepoWebhookCommands.useCreateOne({
        onSuccess: response => {
            toast.success("Webhook created successfully");
            navigateToEdit(response.data);
        },
    });
    const { mutate: updateSettingsRepoWebhook, isPending: isUpdatingSettings } = RepoWebhookCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Webhook updated successfully");
            navigateToList();
        },
    });
    const { mutate: createProjectRepoWebhook, isPending: isCreatingProject } = ProjectRepoWebhookCommands.useCreateOne({
        onSuccess: response => {
            toast.success("Project webhook created successfully");
            navigateToEdit(response.data);
        },
    });
    const { mutate: updateProjectRepoWebhook, isPending: isUpdatingProject } = ProjectRepoWebhookCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project webhook updated successfully");
            navigateToList();
        },
    });

    const settingsDetailQuery = RepoWebhookQueries.useFindOneById(
        { id: detailId },
        {
            enabled: isEditMode && scope.type === "settings" && Boolean(detailId),
        },
    );
    const projectDetailQuery = ProjectRepoWebhookQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        {
            enabled: isEditMode && scope.type === "project" && Boolean(detailId),
        },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingsDetailQuery;
    const repoWebhook = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && repoWebhook?.inherited === true;

    function createPayload(values: CreateOrEditRepoWebhookFormOutput) {
        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            kind: values.kind,
            secret: values.secret,
        };
    }

    function onSubmit(values: CreateOrEditRepoWebhookFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && repoWebhook) {
            const updatePayload = {
                ...payload,
                updateVer: repoWebhook.updateVer,
            };

            if (scope.type === "project") {
                updateProjectRepoWebhook({
                    projectID: scope.projectId,
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

        if (scope.type === "project") {
            createProjectRepoWebhook({
                projectID: scope.projectId,
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

        navigateToList();
    }

    const isPending = isCreatingSettings || isUpdatingSettings || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = scope.type === "settings";
    const matchedCreatedWebhook =
        createdWebhook && isEditMode && createdWebhook.id === repoWebhookId ? createdWebhook : null;
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
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const canRenderForm = mode === "create" || (isEditMode && !!repoWebhook);
    const title = readOnlyInherited ? "Webhook" : mode === "create" ? "Create Webhook" : "Edit Webhook";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && canRenderForm && (
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
        </div>
    );
}

function getRepoWebhookListRoute(scope: RepoWebhookTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.webhooks.$route(scope.projectId);
    }

    return ROUTE.sources.webhooks.$route;
}

interface Props {
    mode: RepoWebhookFormRouteMode;
    scope: RepoWebhookTableScope;
    repoWebhookId?: string;
}
