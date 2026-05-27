import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { ProjectGithubAppQueries } from "~/projects/data/queries";
import { GithubAppCommands } from "~/settings/data/commands";
import { GithubAppQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";

import { CreateOrEditGithubAppForm } from "../form";
import { useCreateOrEditGithubAppDialogState } from "../hooks";
import type { CreateOrEditGithubAppFormOutput } from "../schemas";

export function CreateOrEditGithubAppDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditGithubAppDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");

    const { mutate: createSettingsGithubApp, isPending: isCreatingSettings } = GithubAppCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Github app created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingsGithubApp, isPending: isUpdatingSettings } = GithubAppCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Github app updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectGithubApp, isPending: isCreatingProject } = ProjectGithubAppCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project github app created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectGithubApp, isPending: isUpdatingProject } = ProjectGithubAppCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project github app updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
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

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setTestStatus("idle");
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "edit" ? state.id : "";
    const settingsDetailQuery = GithubAppQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectGithubAppQueries.useFindOneById(
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
    const githubApp = detailQuery.data?.data;

    function createPayload(values: CreateOrEditGithubAppFormOutput) {
        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
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
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values);

        if (state.mode === "edit" && githubApp) {
            const updatePayload = {
                ...payload,
                updateVer: githubApp.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectGithubApp({
                    projectID: state.scope.projectId,
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

        if (state.scope.type === "project") {
            createProjectGithubApp({
                projectID: state.scope.projectId,
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
        if (state.mode !== "edit" || !githubApp) {
            return;
        }

        const payload = {
            name: githubApp.name,
            updateVer: githubApp.updateVer,
        };

        if (state.scope.type === "project") {
            beginProjectReprovision({
                projectID: state.scope.projectId,
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
        ? (resolvedDialogOptions.entityTitle ?? "Github App")
        : "Create or update a Github App";
    const isPending = isCreatingSettings || isUpdatingSettings || isCreatingProject || isUpdatingProject;
    const isReprovisioning = isReprovisioningSettings || isReprovisioningProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const showTestConnection = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues =
        githubApp && state.mode === "edit"
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
        githubApp && state.mode === "edit"
            ? {
                  callbackURL: githubApp.callbackURL,
                  webhookURL: githubApp.webhookURL,
                  webhookSecret: githubApp.webhookSecret,
              }
            : undefined;
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;
    const canRenderForm = state.mode === "open" || (state.mode === "edit" && !!githubApp);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[840px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && canRenderForm && (
                    <CreateOrEditGithubAppForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        isReprovisioning={isReprovisioning}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
                        onReprovision={state.mode === "edit" && !readOnlyInherited ? onReprovision : undefined}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        readonlyValues={readonlyValues}
                        showAvailableInProjects={showAvailableInProjects}
                        showTestConnection={showTestConnection}
                        readOnlyInherited={readOnlyInherited}
                        onClose={handleClose}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
