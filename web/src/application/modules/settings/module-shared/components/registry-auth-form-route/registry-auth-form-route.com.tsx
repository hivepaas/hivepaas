import { useState } from "react";

import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";
import { RegistryAuthCommands } from "~/settings/data/commands";
import { RegistryAuthQueries } from "~/settings/data/queries";
import type { SettingRegistryAuth } from "~/settings/domain";
import { CreateOrEditRegistryAuthForm } from "~/settings/module-shared/components/registry-auth-form";
import type {
    CreateOrEditRegistryAuthFormInput,
    CreateOrEditRegistryAuthFormOutput,
} from "~/settings/module-shared/components/registry-auth-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets, useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import type { RegistryAuthTableScope } from "../registry-auth-table";
import { RevealSecretsButton } from "../reveal-secrets-button";

type RegistryAuthFormRouteMode = "create" | "edit";

export function RegistryAuthFormRoute({ mode, scope, registryAuthId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getRegistryAuthListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (registryAuthId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingRegistryAuth, isPending: isCreatingSetting } = RegistryAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Registry auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingRegistryAuth, isPending: isUpdatingSetting } = RegistryAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Registry auth updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectRegistryAuth, isPending: isCreatingProject } =
        ProjectRegistryAuthCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project registry auth created successfully");
                markSaved();
            },
        });
    const { mutate: updateProjectRegistryAuth, isPending: isUpdatingProject } =
        ProjectRegistryAuthCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project registry auth updated successfully");
                markSaved();
            },
        });
    const { mutate: testConnection, isPending: isTesting } = RegistryAuthCommands.useTestConn({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const settingDetailQuery = RegistryAuthQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectRegistryAuthQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const registryAuth = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && registryAuth?.inherited === true;

    function createPayload(values: CreateOrEditRegistryAuthFormOutput) {
        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            address: values.address,
            username: values.username,
            password: values.password,
            readonly: values.readonly,
        };
    }

    function onSubmit(values: CreateOrEditRegistryAuthFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && registryAuth) {
            const updatePayload = { ...payload, updateVer: registryAuth.updateVer };

            if (scope.type === "project") {
                updateProjectRegistryAuth({
                    projectID: scope.projectId,
                    env: scope.env,
                    id: registryAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingRegistryAuth({ id: registryAuth.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectRegistryAuth({
                projectID: scope.projectId,
                env: scope.env,
                payload,
            });
            return;
        }

        createSettingRegistryAuth({ payload });
    }

    function onTestConnection(values: CreateOrEditRegistryAuthFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                name: values.name,
                address: values.address,
                username: values.username,
                password: values.password,
                readonly: values.readonly,
            },
        });
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
    } = useSettingRevealSecrets<SettingRegistryAuth>({
        settingType: "registry-auth",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || registryAuth?.inherited === true,
        mode,
    });

    const activeRegistryAuth = revealedData ?? registryAuth;

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditRegistryAuthFormInput> | undefined = activeRegistryAuth
        ? {
              name: activeRegistryAuth.name,
              address: activeRegistryAuth.address,
              username: activeRegistryAuth.username,
              password: activeRegistryAuth.password,
              readonly: activeRegistryAuth.readonly,
              inheritable: Boolean(activeRegistryAuth.inheritable),
              default: activeRegistryAuth.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create Registry Auth" : "Edit Registry Auth";

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
                    <CreateOrEditRegistryAuthForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
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

function getRegistryAuthListRoute(scope: RegistryAuthTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.registryAuth.$route(scope.projectId);
    }

    return ROUTE.settings.registryAuth.$route;
}

interface Props {
    mode: RegistryAuthFormRouteMode;
    scope: RegistryAuthTableScope;
    registryAuthId?: string;
}
