import { useMemo, useState } from "react";

import { toast } from "sonner";
import { ProjectBackupRepoCommands } from "~/projects/data/commands";
import { ProjectBackupRepoQueries } from "~/projects/data/queries";
import type { BackupRepo_CreateOne_Payload, BackupRepo_UpdateOne_Payload } from "~/settings/api/services";
import { BackupRepoCommands, BackupRepoQueries } from "~/settings/data";
import type { SettingBackupRepo } from "~/settings/domain";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets, useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import {
    BACKUP_REPO_ACTION,
    BACKUP_REPO_ENGINE,
    BACKUP_REPO_STORAGE_TYPE,
    DEFAULT_BACKUP_REPO_COMPRESSION,
    DEFAULT_BACKUP_REPO_PACK_SIZE,
    DEFAULT_BACKUP_REPO_RETENTION,
} from "../../constants/backup-repo.constants";
import { CreateOrEditBackupRepoForm } from "../backup-repo-form";
import type { CreateOrEditBackupRepoFormInput, CreateOrEditBackupRepoFormOutput } from "../backup-repo-form";
import type { BackupRepoTableScope } from "../backup-repo-table/backup-repo-table.types";
import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";

type BackupRepoFormRouteMode = "create" | "edit";

export function BackupRepoFormRoute({ mode, scope, backupRepoId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getBackupRepoListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (backupRepoId ?? "") : "";
    const projectId = scope.type === "project" ? scope.projectId : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const settingDetailQuery = BackupRepoQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );

    const projectDetailQuery = ProjectBackupRepoQueries.useFindOneById(
        {
            projectID: projectId,
            env: scope.type === "project" ? scope.env : undefined,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );

    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const backupRepo = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && backupRepo?.inherited === true;

    const { mutate: createSettingBackupRepo, isPending: isCreatingSetting } = BackupRepoCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Backup repository created successfully");
            markSaved();
        },
    });

    const { mutate: updateSettingBackupRepo, isPending: isUpdatingSetting } = BackupRepoCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Backup repository updated successfully");
            markSaved();
        },
    });

    const { mutate: createProjectBackupRepo, isPending: isCreatingProject } = ProjectBackupRepoCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project backup repository created successfully");
            markSaved();
        },
    });

    const { mutate: updateProjectBackupRepo, isPending: isUpdatingProject } = ProjectBackupRepoCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project backup repository updated successfully");
            markSaved();
        },
    });

    function onSubmit(values: CreateOrEditBackupRepoFormOutput) {
        if (isEditMode && backupRepo) {
            const updatePayload: BackupRepo_UpdateOne_Payload = {
                updateVer: backupRepo.updateVer,
                name: values.name,
                description: values.description,
                compression: values.compression,
                packSize: values.packSize,
                retention: values.retention,
                inheritable: values.inheritable,
                default: values.default,
            };

            if (scope.type === "project") {
                updateProjectBackupRepo({
                    projectID: scope.projectId,
                    env: scope.env,
                    id: backupRepo.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingBackupRepo({ id: backupRepo.id, payload: updatePayload });
            return;
        }

        const isImport = values.action === BACKUP_REPO_ACTION.ImportExisting;
        const createPayload: BackupRepo_CreateOne_Payload = {
            name: values.name,
            engine: values.engine,
            importExisting: isImport,
            description: isImport ? undefined : values.description,
            cloudStorage:
                values.storageType === BACKUP_REPO_STORAGE_TYPE.CloudStorage && values.cloudStorage?.id
                    ? { id: values.cloudStorage.id }
                    : undefined,
            volume:
                values.storageType === BACKUP_REPO_STORAGE_TYPE.Volume && values.volume?.id
                    ? { id: values.volume.id }
                    : undefined,
            storagePrefix: values.storagePrefix ?? undefined,
            password: values.password,
            compression: isImport ? undefined : values.compression,
            packSize: isImport ? undefined : values.packSize,
            retention: isImport ? undefined : values.retention,
            inheritable: values.inheritable,
            default: values.default,
        };

        if (scope.type === "project") {
            createProjectBackupRepo({
                projectID: scope.projectId,
                env: scope.env,
                payload: createPayload,
            });
            return;
        }

        createSettingBackupRepo({ payload: createPayload });
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
    } = useSettingRevealSecrets<SettingBackupRepo>({
        settingType: "backup-repos",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || backupRepo?.inherited === true,
        mode,
    });

    const activeBackupRepo = revealedData ?? backupRepo;

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;

    const initialValues: Partial<CreateOrEditBackupRepoFormInput> | undefined = useMemo(
        () =>
            isEditMode && activeBackupRepo
                ? {
                      name: activeBackupRepo.name,
                      engine: activeBackupRepo.engine ?? BACKUP_REPO_ENGINE.Kopia,
                      action: BACKUP_REPO_ACTION.CreateNew,
                      description: activeBackupRepo.description ?? "",
                      storageType: activeBackupRepo.volume?.id
                          ? BACKUP_REPO_STORAGE_TYPE.Volume
                          : BACKUP_REPO_STORAGE_TYPE.CloudStorage,
                      cloudStorage: activeBackupRepo.cloudStorage ?? null,
                      volume: activeBackupRepo.volume ?? null,
                      storagePrefix: activeBackupRepo.storagePrefix ?? "",
                      password: activeBackupRepo.password ?? "",
                      compression: activeBackupRepo.compression ?? DEFAULT_BACKUP_REPO_COMPRESSION,
                      packSize: activeBackupRepo.packSize ?? DEFAULT_BACKUP_REPO_PACK_SIZE,
                      retention: {
                          keepLast: activeBackupRepo.retention?.keepLast ?? DEFAULT_BACKUP_REPO_RETENTION.keepLast,
                          keepHourly:
                              activeBackupRepo.retention?.keepHourly ?? DEFAULT_BACKUP_REPO_RETENTION.keepHourly,
                          keepDaily: activeBackupRepo.retention?.keepDaily ?? DEFAULT_BACKUP_REPO_RETENTION.keepDaily,
                          keepWeekly:
                              activeBackupRepo.retention?.keepWeekly ?? DEFAULT_BACKUP_REPO_RETENTION.keepWeekly,
                          keepMonthly:
                              activeBackupRepo.retention?.keepMonthly ?? DEFAULT_BACKUP_REPO_RETENTION.keepMonthly,
                      },
                      inheritable: Boolean(activeBackupRepo.inheritable),
                      default: activeBackupRepo.default ?? false,
                  }
                : undefined,
        [activeBackupRepo, isEditMode],
    );

    const shouldRenderForm = mode === "create" || Boolean(activeBackupRepo);
    const title = mode === "create" ? "Create Backup Repo" : "Edit Backup Repo";

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
                    <CreateOrEditBackupRepoForm
                        key={`${detailId}-${revealRevision}`}
                        backupRepoId={activeBackupRepo?.id ?? detailId}
                        mode={mode}
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        scope={scope}
                        showAvailableInProjects
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                </RevealSecretsProvider>
            )}
        </div>
    );
}

function getBackupRepoListRoute(scope: BackupRepoTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.backupRepos.$route(scope.projectId);
    }

    return ROUTE.settings.backupRepos.$route;
}

interface Props {
    mode: BackupRepoFormRouteMode;
    scope: BackupRepoTableScope;
    backupRepoId?: string;
}
