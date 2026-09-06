import { useState } from "react";

import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { ProjectGithubAppQueries } from "~/projects/data/queries";
import { GithubAppCommands } from "~/settings/data/commands";
import { GithubAppQueries } from "~/settings/data/queries";
import type { SettingGithubApp } from "~/settings/domain";
import { CreateOrEditGithubAppForm } from "~/settings/module-shared/components/github-app-form";
import type {
    CreateOrEditGithubAppFormInput,
    CreateOrEditGithubAppFormOutput,
} from "~/settings/module-shared/components/github-app-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import type { GithubAppTableScope } from "../github-app-table";
import { RevealSecretsButton } from "../reveal-secrets-button";
import { SettingsFormRouteHeader } from "../settings-form-route-header";

type GithubAppFormRouteMode = "create" | "edit";

export function GithubAppFormRoute({ mode, scope, githubAppId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getGithubAppListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (githubAppId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingsGithubApp, isPending: isCreatingSettings } = GithubAppCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Github app created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingsGithubApp, isPending: isUpdatingSettings } = GithubAppCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Github app updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectGithubApp, isPending: isCreatingProject } = ProjectGithubAppCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project github app created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectGithubApp, isPending: isUpdatingProject } = ProjectGithubAppCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project github app updated successfully");
            markSaved();
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
            inheritable: values.inheritable,
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

    const {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    } = useSettingRevealSecrets<SettingGithubApp>({
        settingType: "github-apps",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || githubApp?.inherited === true,
        mode,
    });

    const activeGithubApp = revealedData ?? githubApp;

    const isPending = isCreatingSettings || isUpdatingSettings || isCreatingProject || isUpdatingProject;
    const isReprovisioning = isReprovisioningSettings || isReprovisioningProject;
    const showAvailableInProjects = true;
    const showTestConnection = isEditMode && (!readOnlyInherited || canWrite);
    const initialValues: Partial<CreateOrEditGithubAppFormInput> | undefined =
        isEditMode && activeGithubApp
            ? {
                  name: activeGithubApp.name,
                  organization: activeGithubApp.organization,
                  appId: activeGithubApp.appId,
                  installationId: activeGithubApp.installationId,
                  clientId: activeGithubApp.clientId,
                  clientSecret: activeGithubApp.clientSecret,
                  privateKey: activeGithubApp.privateKey,
                  ssoEnabled: activeGithubApp.ssoEnabled,
                  inheritable: Boolean(activeGithubApp.inheritable),
                  default: activeGithubApp.default ?? false,
              }
            : {
                  ssoEnabled: true,
                  inheritable: scope.type === "project" ? true : false,
                  default: true,
              };
    const readonlyValues =
        activeGithubApp && isEditMode
            ? {
                  callbackURL: activeGithubApp.callbackURL,
                  webhookURL: activeGithubApp.webhookURL,
                  webhookSecret: activeGithubApp.webhookSecret,
              }
            : undefined;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const canRenderForm = mode === "create" || (isEditMode && Boolean(activeGithubApp));
    const title = readOnlyInherited ? "Github App" : mode === "create" ? "Create Github App" : "Edit Github App";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader
                title={title}
                actions={
                    canShowRevealButton ? (
                        <RevealSecretsButton
                            onClick={() => {
                                setIsDialogOpen(true);
                            }}
                            isLoading={isRevealing}
                        />
                    ) : undefined
                }
            />

            <ConfirmRevealSecretsDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                onConfirm={handleConfirmReveal}
                isPending={isRevealing}
            />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && canRenderForm && (
                <RevealSecretsProvider value={{ isRevealed }}>
                    <CreateOrEditGithubAppForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        isReprovisioning={isReprovisioning}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
                        onReprovision={isEditMode && !readOnlyInherited && canWrite ? onReprovision : undefined}
                        settingsURL={activeGithubApp?.settingsURL}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        readonlyValues={readonlyValues}
                        showAvailableInProjects={showAvailableInProjects}
                        isProjectScope={scope.type === "project"}
                        showTestConnection={showTestConnection}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        stickyActions
                        onClose={handleClose}
                    />
                </RevealSecretsProvider>
            )}
        </div>
    );
}

function getGithubAppListRoute(scope: GithubAppTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.githubApps.$route(scope.projectId);
    }

    return ROUTE.settings.githubApps.$route;
}

interface Props {
    mode: GithubAppFormRouteMode;
    scope: GithubAppTableScope;
    githubAppId?: string;
}
