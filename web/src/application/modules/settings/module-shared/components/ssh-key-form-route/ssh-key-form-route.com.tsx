import { useState } from "react";

import { toast } from "sonner";
import { ProjectSSHKeyCommands } from "~/projects/data/commands";
import { ProjectSSHKeyQueries } from "~/projects/data/queries";
import { SSHKeyCommands } from "~/settings/data/commands";
import { SSHKeyQueries } from "~/settings/data/queries";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { CreateOrEditSSHKeyForm } from "~/settings/module-shared/components/ssh-key-form";
import type {
    CreateOrEditSSHKeyFormInput,
    CreateOrEditSSHKeyFormOutput,
} from "~/settings/module-shared/components/ssh-key-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ESSHKeyKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

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
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const sshKey = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && sshKey?.inherited === true;

    function createPayload(values: CreateOrEditSSHKeyFormOutput) {
        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
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
                updateProjectSSHKey({ projectID: scope.projectId, id: sshKey.id, payload: updatePayload });
                return;
            }

            updateSettingSSHKey({ id: sshKey.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectSSHKey({ projectID: scope.projectId, payload });
            return;
        }

        createSettingSSHKey({ payload });
    }

    function handleClose() {
        if (isPending) return;
        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;

        navigateToList();
    }

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditSSHKeyFormInput> | undefined = sshKey
        ? {
              name: sshKey.name,
              kind: getInitialKind(sshKey.kind),
              keyType: sshKey.keyType ?? "",
              publicKey: sshKey.publicKey ?? "",
              privateKey: sshKey.privateKey,
              passphrase: sshKey.passphrase ?? "",
              availableInProjects: sshKey.availableInProjects ?? false,
              default: sshKey.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create SSH Key" : "Edit SSH Key";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditSSHKeyForm
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
                    showAvailableInProjects={scope.type === "settings"}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
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
