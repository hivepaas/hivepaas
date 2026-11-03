import { useState } from "react";

import { toast } from "sonner";
import { ProjectKeyAuthCommands } from "~/projects/data/commands";
import { ProjectKeyAuthQueries } from "~/projects/data/queries";
import { KeyAuthCommands } from "~/settings/data/commands";
import { KeyAuthQueries } from "~/settings/data/queries";
import type { SettingKeyAuth } from "~/settings/domain";
import { CreateOrEditKeyAuthForm } from "~/settings/module-shared/components/key-auth-form";
import type {
    CreateOrEditKeyAuthFormInput,
    CreateOrEditKeyAuthFormOutput,
} from "~/settings/module-shared/components/key-auth-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets, useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import type { KeyAuthTableScope } from "../key-auth-table";
import { RevealSecretsButton } from "../reveal-secrets-button";

type KeyAuthFormRouteMode = "create" | "edit";

export function KeyAuthFormRoute({ mode, scope, keyAuthId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getKeyAuthListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (keyAuthId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const { mutate: createSettingKeyAuth, isPending: isCreatingSetting } = KeyAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Key auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingKeyAuth, isPending: isUpdatingSetting } = KeyAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Key auth updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectKeyAuth, isPending: isCreatingProject } = ProjectKeyAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project key auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectKeyAuth, isPending: isUpdatingProject } = ProjectKeyAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project key auth updated successfully");
            markSaved();
        },
    });

    const settingDetailQuery = KeyAuthQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectKeyAuthQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const keyAuth = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && keyAuth?.inherited === true;

    function createPayload(values: CreateOrEditKeyAuthFormOutput) {
        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            keyId: values.keyId,
            secretKey: values.secretKey,
        };
    }

    function onSubmit(values: CreateOrEditKeyAuthFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && keyAuth) {
            const updatePayload = { ...payload, updateVer: keyAuth.updateVer };

            if (scope.type === "project") {
                updateProjectKeyAuth({
                    projectID: scope.projectId,
                    env: scope.env,
                    id: keyAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingKeyAuth({ id: keyAuth.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectKeyAuth({
                projectID: scope.projectId,
                env: scope.env,
                payload,
            });
            return;
        }

        createSettingKeyAuth({ payload });
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
    } = useSettingRevealSecrets<SettingKeyAuth>({
        settingType: "key-auth",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || keyAuth?.inherited === true,
        mode,
    });

    const activeKeyAuth = revealedData ?? keyAuth;

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditKeyAuthFormInput> | undefined = activeKeyAuth
        ? {
              name: activeKeyAuth.name,
              keyId: activeKeyAuth.keyId,
              secretKey: activeKeyAuth.secretKey,
              inheritable: Boolean(activeKeyAuth.inheritable),
              default: activeKeyAuth.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create Key Auth" : "Edit Key Auth";

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
                    <CreateOrEditKeyAuthForm
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

function getKeyAuthListRoute(scope: KeyAuthTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.keyAuth.$route(scope.projectId);
    }

    return ROUTE.settings.keyAuth.$route;
}

interface Props {
    mode: KeyAuthFormRouteMode;
    scope: KeyAuthTableScope;
    keyAuthId?: string;
}
