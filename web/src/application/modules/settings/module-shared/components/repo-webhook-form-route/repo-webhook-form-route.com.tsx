import { useState } from "react";

import { useLocation } from "react-router";
import { toast } from "sonner";
import { ProjectRepoWebhookCommands } from "~/projects/data/commands";
import { ProjectRepoWebhookQueries } from "~/projects/data/queries";
import { RepoWebhookCommands } from "~/settings/data/commands";
import { RepoWebhookQueries } from "~/settings/data/queries";
import { CreateOrEditRepoWebhookForm } from "~/settings/module-shared/components/repo-webhook-form";
import type { CreateOrEditRepoWebhookFormOutput } from "~/settings/module-shared/components/repo-webhook-form";
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
    const [saveRevision, setSaveRevision] = useState(0);
    const location = useLocation();
    const [createdWebhook, setCreatedWebhook] = useState<CreatedWebhookResult | null>(
        () => (location.state as RepoWebhookLocationState | null)?.createdRepoWebhook ?? null,
    );
    const [createdWebhookValues, setCreatedWebhookValues] = useState<CreateOrEditRepoWebhookFormOutput | null>(null);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (repoWebhookId ?? "") : "";

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingsRepoWebhook, isPending: isCreatingSettings } = RepoWebhookCommands.useCreateOne({
        onSuccess: response => {
            toast.success("Webhook created successfully");
            setCreatedWebhook(response.data);
            markSaved();
        },
    });
    const { mutate: updateSettingsRepoWebhook, isPending: isUpdatingSettings } = RepoWebhookCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Webhook updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectRepoWebhook, isPending: isCreatingProject } = ProjectRepoWebhookCommands.useCreateOne({
        onSuccess: response => {
            toast.success("Project webhook created successfully");
            setCreatedWebhook(response.data);
            markSaved();
        },
    });
    const { mutate: updateProjectRepoWebhook, isPending: isUpdatingProject } = ProjectRepoWebhookCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project webhook updated successfully");
            markSaved();
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
            setCreatedWebhookValues(values);
            createProjectRepoWebhook({
                projectID: scope.projectId,
                payload,
            });
            return;
        }

        setCreatedWebhookValues(values);
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

        navigate.modules(getRepoWebhookListRoute(scope), { ignorePrevPath: true });
    }

    const isPending = isCreatingSettings || isUpdatingSettings || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = scope.type === "settings";
    const matchedCreatedWebhook =
        createdWebhook && isEditMode && createdWebhook.id === repoWebhookId ? createdWebhook : null;
    const createdInitialValues =
        !isEditMode && createdWebhook && createdWebhookValues
            ? {
                  ...createdWebhookValues,
                  secret: createdWebhook.secret,
              }
            : undefined;
    const initialValues = repoWebhook
        ? {
              name: repoWebhook.name,
              kind: repoWebhook.kind as ERepoWebhookKind,
              secret: matchedCreatedWebhook?.secret ?? repoWebhook.secret,
              availableInProjects: repoWebhook.availableInProjects ?? false,
              default: repoWebhook.default ?? false,
          }
        : createdInitialValues;
    const webhookURL = repoWebhook?.webhookURL ?? matchedCreatedWebhook?.webhookURL ?? createdWebhook?.webhookURL;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const canRenderForm = mode === "create" || (isEditMode && !!repoWebhook);
    const title = readOnlyInherited ? "Webhook" : mode === "create" ? "Create Webhook" : "Edit Webhook";

    return (
        <div className="flex w-full flex-col">
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
                    savedVersion={saveRevision}
                    key={createdWebhook?.id ?? "new"}
                    initialValues={initialValues}
                    webhookURL={webhookURL}
                    showAvailableInProjects={showAvailableInProjects}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    stickyActions
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function getRepoWebhookListRoute(scope: RepoWebhookTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.sources.webhooks.$route(scope.projectId);
    }

    return ROUTE.sources.webhooks.$route;
}

interface Props {
    mode: RepoWebhookFormRouteMode;
    scope: RepoWebhookTableScope;
    repoWebhookId?: string;
}
