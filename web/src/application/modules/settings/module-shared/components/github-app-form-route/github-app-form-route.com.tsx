import { useState } from "react";

import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { ProjectGithubAppQueries } from "~/projects/data/queries";
import { GithubAppCommands } from "~/settings/data/commands";
import { GithubAppQueries } from "~/settings/data/queries";
import { CreateOrEditGithubAppForm } from "~/settings/dialogs/create-or-edit-github-app/form";
import type { CreateOrEditGithubAppFormOutput } from "~/settings/dialogs/create-or-edit-github-app/schemas";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { GithubAppTableScope } from "../github-app-table";
import { SettingsFormRouteHeader } from "../settings-form-route-header";

type GithubAppFormRouteMode = "create" | "edit";

export function GithubAppFormRoute({ mode, scope, githubAppId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getGithubAppListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (githubAppId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    const { mutate: createSettingsGithubApp, isPending: isCreatingSettings } = GithubAppCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Github app created successfully");
            navigateToList();
        },
    });
    const { mutate: updateSettingsGithubApp, isPending: isUpdatingSettings } = GithubAppCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Github app updated successfully");
            navigateToList();
        },
    });
    const { mutate: createProjectGithubApp, isPending: isCreatingProject } = ProjectGithubAppCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project github app created successfully");
            navigateToList();
        },
    });
    const { mutate: updateProjectGithubApp, isPending: isUpdatingProject } = ProjectGithubAppCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project github app updated successfully");
            navigateToList();
        },
    });
    const { mutate: testConnection, isPending: isTesting } = GithubAppCommands.useTestConnection({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });
    const { mutate: beginSettingsReprovision, isPending: isReprovisioningSettings } =
        GithubAppCommands.useBeginReprovision({
            onSuccess: response => {
                window.location.assign(response.data.redirectURL);
            },
        });
    const { mutate: beginProjectReprovision, isPending: isReprovisioningProject } =
        ProjectGithubAppCommands.useBeginReprovision({
            onSuccess: response => {
                window.location.assign(response.data.redirectURL);
            },
        });

    const settingsDetailQuery = GithubAppQueries.useFindOneById(
        { id: detailId },
        {
            enabled: isEditMode && scope.type === "settings" && Boolean(detailId),
        },
    );
    const projectDetailQuery = ProjectGithubAppQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        {
            enabled: isEditMode && scope.type === "project" && Boolean(detailId),
        },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingsDetailQuery;
    const githubApp = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && githubApp?.inherited === true;

    function createPayload(values: CreateOrEditGithubAppFormOutput) {
        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            organization: values.organization,
            appId: values.appId,
            installationId: values.installationId,
            clientId: values.clientId,
            clientSecret: values.clientSecret,
            privateKey: values.privateKey,
            ssoEnabled: values.ssoEnabled,
        };
    }

    function onSubmit(values: CreateOrEditGithubAppFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && githubApp) {
            const updatePayload = {
                ...payload,
                updateVer: githubApp.updateVer,
            };

            if (scope.type === "project") {
                updateProjectGithubApp({
                    projectID: scope.projectId,
                    id: githubApp.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingsGithubApp({
                id: githubApp.id,
                payload: updatePayload,
            });
            return;
        }

        if (scope.type === "project") {
            createProjectGithubApp({
                projectID: scope.projectId,
                payload,
            });
            return;
        }

        createSettingsGithubApp({ payload });
    }

    function onTestConnection(values: CreateOrEditGithubAppFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                name: values.name,
                organization: values.organization,
                appId: values.appId,
                installationId: values.installationId,
                clientId: values.clientId,
                clientSecret: values.clientSecret,
                privateKey: values.privateKey,
                ssoEnabled: values.ssoEnabled,
            },
        });
    }

    function onReprovision() {
        if (!isEditMode || !githubApp) {
            return;
        }

        const payload = {
            name: githubApp.name,
            updateVer: githubApp.updateVer,
        };

        if (scope.type === "project") {
            beginProjectReprovision({
                projectID: scope.projectId,
                id: githubApp.id,
                payload,
            });
            return;
        }

        beginSettingsReprovision({
            id: githubApp.id,
            payload,
        });
    }

    function handleClose() {
        if (isPending || isTesting || isReprovisioning) {
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
    const isReprovisioning = isReprovisioningSettings || isReprovisioningProject;
    const showAvailableInProjects = scope.type === "settings";
    const showTestConnection = scope.type === "settings";
    const initialValues =
        githubApp && isEditMode
            ? {
                  name: githubApp.name,
                  organization: githubApp.organization,
                  appId: githubApp.appId || undefined,
                  installationId: githubApp.installationId || undefined,
                  clientId: githubApp.clientId,
                  clientSecret: githubApp.secretMasked ? "" : githubApp.clientSecret,
                  privateKey: githubApp.secretMasked ? "" : githubApp.privateKey,
                  ssoEnabled: githubApp.ssoEnabled,
                  availableInProjects: githubApp.availableInProjects ?? false,
                  default: githubApp.default ?? false,
              }
            : {
                  ssoEnabled: true,
                  availableInProjects: true,
                  default: true,
              };
    const readonlyValues =
        githubApp && isEditMode
            ? {
                  callbackURL: githubApp.callbackURL,
                  webhookURL: githubApp.webhookURL,
                  webhookSecret: githubApp.webhookSecret,
              }
            : undefined;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const canRenderForm = mode === "create" || (isEditMode && !!githubApp);
    const title = readOnlyInherited ? "Github App" : mode === "create" ? "Create Github App" : "Edit Github App";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && canRenderForm && (
                <CreateOrEditGithubAppForm
                    isPending={isPending}
                    isTesting={isTesting}
                    testStatus={testStatus}
                    isReprovisioning={isReprovisioning}
                    onSubmit={onSubmit}
                    onTestConnection={onTestConnection}
                    onReprovision={isEditMode && !readOnlyInherited && canWrite ? onReprovision : undefined}
                    onHasChanges={setHasChanges}
                    initialValues={initialValues}
                    readonlyValues={readonlyValues}
                    showAvailableInProjects={showAvailableInProjects}
                    showTestConnection={showTestConnection}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function getGithubAppListRoute(scope: GithubAppTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.sources.githubApps.$route(scope.projectId);
    }

    return ROUTE.sources.githubApps.$route;
}

interface Props {
    mode: GithubAppFormRouteMode;
    scope: GithubAppTableScope;
    githubAppId?: string;
}
