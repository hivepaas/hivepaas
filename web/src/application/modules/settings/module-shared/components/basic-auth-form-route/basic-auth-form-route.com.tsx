import { useState } from "react";

import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { ProjectBasicAuthQueries } from "~/projects/data/queries";
import { BasicAuthCommands } from "~/settings/data/commands";
import { BasicAuthQueries } from "~/settings/data/queries";
import type { SettingBasicAuth } from "~/settings/domain";
import { CreateOrEditBasicAuthForm } from "~/settings/module-shared/components/basic-auth-form";
import type {
    CreateOrEditBasicAuthFormInput,
    CreateOrEditBasicAuthFormOutput,
} from "~/settings/module-shared/components/basic-auth-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets, useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import type { BasicAuthTableScope } from "../basic-auth-table";
import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";

type BasicAuthFormRouteMode = "create" | "edit";

export function BasicAuthFormRoute({ mode, scope, basicAuthId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getBasicAuthListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (basicAuthId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingBasicAuth, isPending: isCreatingSetting } = BasicAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Basic auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingBasicAuth, isPending: isUpdatingSetting } = BasicAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Basic auth updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectBasicAuth, isPending: isCreatingProject } = ProjectBasicAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project basic auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectBasicAuth, isPending: isUpdatingProject } = ProjectBasicAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project basic auth updated successfully");
            markSaved();
        },
    });

    const settingDetailQuery = BasicAuthQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectBasicAuthQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const basicAuth = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && basicAuth?.inherited === true;

    function createPayload(values: CreateOrEditBasicAuthFormOutput) {
        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            username: values.username,
            password: values.password,
        };
    }

    function onSubmit(values: CreateOrEditBasicAuthFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && basicAuth) {
            const updatePayload = { ...payload, updateVer: basicAuth.updateVer };

            if (scope.type === "project") {
                updateProjectBasicAuth({
                    projectID: scope.projectId,
                    env: scope.env,
                    id: basicAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingBasicAuth({ id: basicAuth.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectBasicAuth({
                projectID: scope.projectId,
                env: scope.env,
                payload,
            });
            return;
        }

        createSettingBasicAuth({ payload });
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

    const {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    } = useSettingRevealSecrets<SettingBasicAuth>({
        settingType: "basic-auth",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || basicAuth?.inherited === true,
        mode,
    });

    const activeBasicAuth = revealedData ?? basicAuth;

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditBasicAuthFormInput> | undefined = activeBasicAuth
        ? {
              name: activeBasicAuth.name,
              username: activeBasicAuth.username,
              password: activeBasicAuth.password,
              inheritable: Boolean(activeBasicAuth.inheritable),
              default: activeBasicAuth.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create Basic Auth" : "Edit Basic Auth";

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

            {!isDetailLoading && shouldRenderForm && (
                <RevealSecretsProvider value={{ isRevealed }}>
                    <CreateOrEditBasicAuthForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        showAvailableInProjects
                        isProjectScope={scope.type === "project"}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                </RevealSecretsProvider>
            )}
        </div>
    );
}

function getBasicAuthListRoute(scope: BasicAuthTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.basicAuth.$route(scope.projectId);
    }

    return ROUTE.settings.basicAuth.$route;
}

interface Props {
    mode: BasicAuthFormRouteMode;
    scope: BasicAuthTableScope;
    basicAuthId?: string;
}
