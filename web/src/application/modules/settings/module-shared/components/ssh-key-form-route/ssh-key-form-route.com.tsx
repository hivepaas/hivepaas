import { useState } from "react";

import { toast } from "sonner";
import { ProjectSSHKeyCommands } from "~/projects/data/commands";
import { ProjectSSHKeyQueries } from "~/projects/data/queries";
import { SSHKeyCommands } from "~/settings/data/commands";
import { SSHKeyQueries } from "~/settings/data/queries";
import type { SettingSSHKey } from "~/settings/domain";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { CreateOrEditSSHKeyForm } from "~/settings/module-shared/components/ssh-key-form";
import type {
    CreateOrEditSSHKeyFormInput,
    CreateOrEditSSHKeyFormOutput,
} from "~/settings/module-shared/components/ssh-key-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ESSHKeyKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";
import type { SSHKeyTableScope } from "../ssh-key-table";

type SSHKeyFormRouteMode = "create" | "edit";
const SSH_KEY_KIND_VALUES = Object.values(ESSHKeyKind);

export function SSHKeyFormRoute({ mode, scope, sshKeyId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getSSHKeyListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (sshKeyId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingSSHKey, isPending: isCreatingSetting } = SSHKeyCommands.useCreateOne({
        onSuccess: () => {
            toast.success("SSH key created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingSSHKey, isPending: isUpdatingSetting } = SSHKeyCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSH key updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectSSHKey, isPending: isCreatingProject } = ProjectSSHKeyCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project SSH key created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectSSHKey, isPending: isUpdatingProject } = ProjectSSHKeyCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project SSH key updated successfully");
            markSaved();
        },
    });
    const { mutateAsync: generateSSHKey, isPending: isGenerating } = SSHKeyCommands.useGenerate();

    const settingDetailQuery = SSHKeyQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectSSHKeyQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const sshKey = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && sshKey?.inherited === true;

    function createPayload(values: CreateOrEditSSHKeyFormOutput) {
        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            kind: values.kind,
            keyType: values.keyType,
            publicKey: values.publicKey,
            privateKey: values.privateKey,
            passphrase: values.passphrase,
        };
    }

    function onSubmit(values: CreateOrEditSSHKeyFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && sshKey) {
            const updatePayload = { ...payload, updateVer: sshKey.updateVer };

            if (scope.type === "project") {
                updateProjectSSHKey({
                    projectID: scope.projectId,
                    env: scope.env,
                    id: sshKey.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingSSHKey({ id: sshKey.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectSSHKey({
                projectID: scope.projectId,
                env: scope.env,
                payload,
            });
            return;
        }

        createSettingSSHKey({ payload });
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
    } = useSettingRevealSecrets<SettingSSHKey>({
        settingType: "ssh-keys",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || sshKey?.inherited === true,
        mode,
    });

    const activeSSHKey = revealedData ?? sshKey;

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditSSHKeyFormInput> | undefined = activeSSHKey
        ? {
              name: activeSSHKey.name,
              kind: getInitialKind(activeSSHKey.kind),
              keyType: activeSSHKey.keyType ?? "",
              publicKey: activeSSHKey.publicKey ?? "",
              privateKey: activeSSHKey.privateKey,
              passphrase: activeSSHKey.passphrase ?? "",
              inheritable: Boolean(activeSSHKey.inheritable),
              default: activeSSHKey.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create SSH Key" : "Edit SSH Key";

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
                    <CreateOrEditSSHKeyForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        isGenerating={isGenerating}
                        onGenerate={async payload => {
                            const response = await generateSSHKey({ payload });
                            return response.data;
                        }}
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

function getInitialKind(kind?: string): ESSHKeyKind {
    if (SSH_KEY_KIND_VALUES.includes(kind as ESSHKeyKind)) {
        return kind as ESSHKeyKind;
    }

    return ESSHKeyKind.Git;
}

function getSSHKeyListRoute(scope: SSHKeyTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.sshKeys.$route(scope.projectId);
    }

    return ROUTE.settings.sshKeys.$route;
}

interface Props {
    mode: SSHKeyFormRouteMode;
    scope: SSHKeyTableScope;
    sshKeyId?: string;
}
