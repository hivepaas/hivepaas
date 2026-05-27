import { useEffect, useMemo, useState } from "react";

import type { ColumnDef } from "@tanstack/react-table";
import { toast } from "sonner";
import {
    PROJECT_SETTINGS_IMPORT_KIND,
    ProjectSettingsImportCommands,
    type ProjectSettingsImportKind,
} from "~/projects/data/commands";
import {
    ProjectAccessTokenQueries,
    ProjectBasicAuthQueries,
    ProjectCloudStorageQueries,
    ProjectEmailQueries,
    ProjectGithubAppQueries,
    ProjectImServiceQueries,
    ProjectNotificationQueries,
    ProjectRegistryAuthQueries,
    ProjectSSHKeyQueries,
    ProjectSslCertQueries,
} from "~/projects/data/queries";
import {
    AccessTokenQueries,
    BasicAuthQueries,
    CloudStorageQueries,
    EmailQueries,
    GithubAppQueries,
    ImServiceQueries,
    NotificationQueries,
    RegistryAuthQueries,
    SSHKeyQueries,
    SslCertQueries,
} from "~/settings/data/queries";
import type { SettingsBaseEntity } from "~/settings/domain";
import { AccessTokenTableDefs } from "~/settings/module-shared/components/access-token-table/access-token-table.defs";
import { BasicAuthTableDefs } from "~/settings/module-shared/components/basic-auth-table/basic-auth-table.defs";
import { CloudStorageTableDefs } from "~/settings/module-shared/components/cloud-storage-table/cloud-storage-table.defs";
import { EmailAccountTableDefs } from "~/settings/module-shared/components/email-account-table/email-account-table.defs";
import { GithubAppTableDefs } from "~/settings/module-shared/components/github-app-table/github-app-table.defs";
import { ImPlatformTableDefs } from "~/settings/module-shared/components/im-platform-table/im-platform-table.defs";
import { NotificationTargetTableDefs } from "~/settings/module-shared/components/notification-target-table/notification-target-table.defs";
import { RegistryAuthTableDefs } from "~/settings/module-shared/components/registry-auth-table/registry-auth-table.defs";
import { SSHKeyTableDefs } from "~/settings/module-shared/components/ssh-key-table/ssh-key-table.defs";
import { SslCertTableDefs } from "~/settings/module-shared/components/ssl-cert-table/ssl-cert-table.defs";

import { TableActions } from "@application/shared/components";
import { useTableState } from "@application/shared/hooks/table";

import { Button, Checkbox, DataTable } from "@/components/ui";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";

import { useImportProjectSettingsDialogState } from "../hooks";

const IMPORT_DIALOG_PAGINATION = { page: 1, size: 10000 };

type ImportableSetting = SettingsBaseEntity & {
    inherited?: boolean;
};

const EMPTY_IMPORT_SETTINGS: ImportableSetting[] = [];

const IMPORT_DIALOG_LABELS = {
    [PROJECT_SETTINGS_IMPORT_KIND.BasicAuth]: "Basic Auth",
    [PROJECT_SETTINGS_IMPORT_KIND.RegistryAuth]: "Registry Auth",
    [PROJECT_SETTINGS_IMPORT_KIND.SslCert]: "SSL Certificates",
    [PROJECT_SETTINGS_IMPORT_KIND.Email]: "Email Accounts",
    [PROJECT_SETTINGS_IMPORT_KIND.ImService]: "IM Platforms",
    [PROJECT_SETTINGS_IMPORT_KIND.SSHKey]: "SSH Keys",
    [PROJECT_SETTINGS_IMPORT_KIND.AccessToken]: "Access Tokens",
    [PROJECT_SETTINGS_IMPORT_KIND.CloudStorage]: "Cloud Storages",
    [PROJECT_SETTINGS_IMPORT_KIND.Notification]: "Notification Targets",
    [PROJECT_SETTINGS_IMPORT_KIND.GithubApp]: "Github Apps",
} as const satisfies Record<ProjectSettingsImportKind, string>;

function castColumns<T extends ImportableSetting>(columns: ColumnDef<T>[]): ColumnDef<ImportableSetting>[] {
    return columns as unknown as ColumnDef<ImportableSetting>[];
}

function getColumnId(column: ColumnDef<ImportableSetting>) {
    if ("id" in column && column.id) {
        return column.id;
    }

    if ("accessorKey" in column && column.accessorKey) {
        return String(column.accessorKey);
    }

    return "";
}

function getImportColumns(settingKind: ProjectSettingsImportKind | null): ColumnDef<ImportableSetting>[] {
    switch (settingKind) {
        case PROJECT_SETTINGS_IMPORT_KIND.BasicAuth:
            return castColumns(BasicAuthTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.RegistryAuth:
            return castColumns(RegistryAuthTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.SslCert:
            return castColumns(SslCertTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.Email:
            return castColumns(EmailAccountTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.ImService:
            return castColumns(ImPlatformTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.SSHKey:
            return castColumns(SSHKeyTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.AccessToken:
            return castColumns(AccessTokenTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.CloudStorage:
            return castColumns(CloudStorageTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.Notification:
            return castColumns(NotificationTargetTableDefs.columns({ type: "settings" }));
        case PROJECT_SETTINGS_IMPORT_KIND.GithubApp:
            return castColumns(GithubAppTableDefs.columns({ type: "settings" }));
        default:
            return [];
    }
}

function filterImportColumns(columns: ColumnDef<ImportableSetting>[]) {
    return columns.filter(column => {
        const columnId = getColumnId(column);

        return columnId !== "view" && columnId !== "actions";
    });
}

export function ImportProjectSettingsDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useImportProjectSettingsDialogState();
    const { sorting, setSorting, search, setSearch } = useTableState();
    const [selectedIds, setSelectedIds] = useState<Set<string>>(() => new Set());

    const open = state.mode !== "closed";
    const projectId = state.mode === "open" ? state.projectId : "";
    const settingKind = state.mode === "open" ? state.settingKind : null;

    const queryRequest = {
        pagination: IMPORT_DIALOG_PAGINATION,
        sorting,
        search,
    };
    const projectListRequest = {
        projectID: projectId,
        pagination: IMPORT_DIALOG_PAGINATION,
    };

    const basicAuthSettingsQuery = BasicAuthQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.BasicAuth,
    });
    const basicAuthProjectQuery = ProjectBasicAuthQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.BasicAuth,
    });

    const registryAuthSettingsQuery = RegistryAuthQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.RegistryAuth,
    });
    const registryAuthProjectQuery = ProjectRegistryAuthQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.RegistryAuth,
    });

    const sslCertSettingsQuery = SslCertQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.SslCert,
    });
    const sslCertProjectQuery = ProjectSslCertQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.SslCert,
    });

    const emailSettingsQuery = EmailQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.Email,
    });
    const emailProjectQuery = ProjectEmailQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.Email,
    });

    const imServiceSettingsQuery = ImServiceQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.ImService,
    });
    const imServiceProjectQuery = ProjectImServiceQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.ImService,
    });

    const sshKeySettingsQuery = SSHKeyQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.SSHKey,
    });
    const sshKeyProjectQuery = ProjectSSHKeyQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.SSHKey,
    });

    const accessTokenSettingsQuery = AccessTokenQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.AccessToken,
    });
    const accessTokenProjectQuery = ProjectAccessTokenQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.AccessToken,
    });

    const cloudStorageSettingsQuery = CloudStorageQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.CloudStorage,
    });
    const cloudStorageProjectQuery = ProjectCloudStorageQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.CloudStorage,
    });

    const notificationSettingsQuery = NotificationQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.Notification,
    });
    const notificationProjectQuery = ProjectNotificationQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.Notification,
    });

    const githubAppSettingsQuery = GithubAppQueries.useFindManyPaginated(queryRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.GithubApp,
    });
    const githubAppProjectQuery = ProjectGithubAppQueries.useFindManyPaginated(projectListRequest, {
        enabled: open && settingKind === PROJECT_SETTINGS_IMPORT_KIND.GithubApp,
    });

    let settings: ImportableSetting[] = EMPTY_IMPORT_SETTINGS;
    let projectSettings: ImportableSetting[] = EMPTY_IMPORT_SETTINGS;
    let isFetching = false;
    let hasError = false;
    let refetch: () => void = () => undefined;

    switch (settingKind) {
        case PROJECT_SETTINGS_IMPORT_KIND.BasicAuth:
            settings = basicAuthSettingsQuery.data?.data ?? [];
            projectSettings = basicAuthProjectQuery.data?.data ?? [];
            isFetching = basicAuthSettingsQuery.isFetching || basicAuthProjectQuery.isFetching;
            hasError = Boolean(basicAuthSettingsQuery.error ?? basicAuthProjectQuery.error);
            refetch = () => {
                void basicAuthSettingsQuery.refetch();
                void basicAuthProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.RegistryAuth:
            settings = registryAuthSettingsQuery.data?.data ?? [];
            projectSettings = registryAuthProjectQuery.data?.data ?? [];
            isFetching = registryAuthSettingsQuery.isFetching || registryAuthProjectQuery.isFetching;
            hasError = Boolean(registryAuthSettingsQuery.error ?? registryAuthProjectQuery.error);
            refetch = () => {
                void registryAuthSettingsQuery.refetch();
                void registryAuthProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.SslCert:
            settings = sslCertSettingsQuery.data?.data ?? [];
            projectSettings = sslCertProjectQuery.data?.data ?? [];
            isFetching = sslCertSettingsQuery.isFetching || sslCertProjectQuery.isFetching;
            hasError = Boolean(sslCertSettingsQuery.error ?? sslCertProjectQuery.error);
            refetch = () => {
                void sslCertSettingsQuery.refetch();
                void sslCertProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.Email:
            settings = emailSettingsQuery.data?.data ?? [];
            projectSettings = emailProjectQuery.data?.data ?? [];
            isFetching = emailSettingsQuery.isFetching || emailProjectQuery.isFetching;
            hasError = Boolean(emailSettingsQuery.error ?? emailProjectQuery.error);
            refetch = () => {
                void emailSettingsQuery.refetch();
                void emailProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.ImService:
            settings = imServiceSettingsQuery.data?.data ?? [];
            projectSettings = imServiceProjectQuery.data?.data ?? [];
            isFetching = imServiceSettingsQuery.isFetching || imServiceProjectQuery.isFetching;
            hasError = Boolean(imServiceSettingsQuery.error ?? imServiceProjectQuery.error);
            refetch = () => {
                void imServiceSettingsQuery.refetch();
                void imServiceProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.SSHKey:
            settings = sshKeySettingsQuery.data?.data ?? [];
            projectSettings = sshKeyProjectQuery.data?.data ?? [];
            isFetching = sshKeySettingsQuery.isFetching || sshKeyProjectQuery.isFetching;
            hasError = Boolean(sshKeySettingsQuery.error ?? sshKeyProjectQuery.error);
            refetch = () => {
                void sshKeySettingsQuery.refetch();
                void sshKeyProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.AccessToken:
            settings = accessTokenSettingsQuery.data?.data ?? [];
            projectSettings = accessTokenProjectQuery.data?.data ?? [];
            isFetching = accessTokenSettingsQuery.isFetching || accessTokenProjectQuery.isFetching;
            hasError = Boolean(accessTokenSettingsQuery.error ?? accessTokenProjectQuery.error);
            refetch = () => {
                void accessTokenSettingsQuery.refetch();
                void accessTokenProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.CloudStorage:
            settings = cloudStorageSettingsQuery.data?.data ?? [];
            projectSettings = cloudStorageProjectQuery.data?.data ?? [];
            isFetching = cloudStorageSettingsQuery.isFetching || cloudStorageProjectQuery.isFetching;
            hasError = Boolean(cloudStorageSettingsQuery.error ?? cloudStorageProjectQuery.error);
            refetch = () => {
                void cloudStorageSettingsQuery.refetch();
                void cloudStorageProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.Notification:
            settings = notificationSettingsQuery.data?.data ?? [];
            projectSettings = notificationProjectQuery.data?.data ?? [];
            isFetching = notificationSettingsQuery.isFetching || notificationProjectQuery.isFetching;
            hasError = Boolean(notificationSettingsQuery.error ?? notificationProjectQuery.error);
            refetch = () => {
                void notificationSettingsQuery.refetch();
                void notificationProjectQuery.refetch();
            };
            break;
        case PROJECT_SETTINGS_IMPORT_KIND.GithubApp:
            settings = githubAppSettingsQuery.data?.data ?? [];
            projectSettings = githubAppProjectQuery.data?.data ?? [];
            isFetching = githubAppSettingsQuery.isFetching || githubAppProjectQuery.isFetching;
            hasError = Boolean(githubAppSettingsQuery.error ?? githubAppProjectQuery.error);
            refetch = () => {
                void githubAppSettingsQuery.refetch();
                void githubAppProjectQuery.refetch();
            };
            break;
    }

    const inheritedSettingIds = useMemo(
        () => new Set(projectSettings.filter(setting => setting.inherited).map(setting => setting.id)),
        [projectSettings],
    );

    const importableSettings = useMemo(
        () => settings.filter(setting => !inheritedSettingIds.has(setting.id)),
        [inheritedSettingIds, settings],
    );

    const importableSettingIdsKey = useMemo(
        () => importableSettings.map(setting => setting.id).join("|"),
        [importableSettings],
    );

    const allVisibleSelected =
        importableSettings.length > 0 && importableSettings.every(setting => selectedIds.has(setting.id));
    const someVisibleSelected = importableSettings.some(setting => selectedIds.has(setting.id));

    const selectionColumn = useMemo<ColumnDef<ImportableSetting>>(
        () => ({
            id: "select",
            header: () => (
                <Checkbox
                    checked={allVisibleSelected ? true : someVisibleSelected ? "indeterminate" : false}
                    disabled={importableSettings.length === 0}
                    onCheckedChange={checked => {
                        const shouldSelect = checked === true;
                        setSelectedIds(current => {
                            const next = new Set(current);
                            for (const setting of importableSettings) {
                                if (shouldSelect) {
                                    next.add(setting.id);
                                } else {
                                    next.delete(setting.id);
                                }
                            }
                            return next;
                        });
                    }}
                />
            ),
            enableSorting: false,
            enableHiding: false,
            size: 48,
            cell: ({ row: { original } }) => (
                <Checkbox
                    checked={selectedIds.has(original.id)}
                    onCheckedChange={checked => {
                        setSelectedIds(current => {
                            const next = new Set(current);
                            if (checked === true) {
                                next.add(original.id);
                            } else {
                                next.delete(original.id);
                            }
                            return next;
                        });
                    }}
                />
            ),
            meta: {
                align: "center",
                titleAlign: "center",
            },
        }),
        [allVisibleSelected, importableSettings, selectedIds, someVisibleSelected],
    );

    const columns = useMemo(
        () => [selectionColumn, ...filterImportColumns(getImportColumns(settingKind))],
        [selectionColumn, settingKind],
    );

    const { mutate: importSettings, isPending } = ProjectSettingsImportCommands.useImportSettings({
        onSuccess: () => {
            toast.success("Settings imported successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    useEffect(() => {
        if (open) {
            return;
        }

        setSelectedIds(new Set());
        setSearch("");
        setSorting([]);
        clearDialog();
    }, [clearDialog, open, setSearch, setSorting]);

    useEffect(() => {
        setSelectedIds(current => {
            const importableIds = new Set(importableSettings.map(setting => setting.id));
            const next = new Set([...current].filter(id => importableIds.has(id)));

            if (next.size === current.size) {
                return current;
            }

            return next;
        });
    }, [importableSettingIdsKey, importableSettings]);

    function handleClose() {
        if (isPending) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    function handleImport() {
        if (state.mode !== "open" || selectedIds.size === 0) {
            return;
        }

        importSettings({
            projectID: state.projectId,
            settingKind: state.settingKind,
            payload: {
                settings: [...selectedIds].map(id => ({ id })),
                dataViewAllowed: false,
            },
        });
    }

    const title = settingKind ? IMPORT_DIALOG_LABELS[settingKind] : "Settings";

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogContent className="min-w-[390px] w-[980px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Import {title}</DialogTitle>
                </DialogHeader>

                {hasError ? (
                    <div className="flex min-h-48 flex-col items-center justify-center gap-3 text-sm text-muted-foreground">
                        <span>Unable to load settings.</span>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={refetch}
                        >
                            Retry
                        </Button>
                    </div>
                ) : (
                    <div className="flex flex-col gap-4">
                        <TableActions search={{ value: search, onChange: setSearch }} />
                        <DataTable
                            columns={columns}
                            data={importableSettings}
                            pageSize={10}
                            manualSorting
                            enableSorting
                            enablePagination
                            isLoading={isFetching}
                            onSortingChange={setSorting}
                            showPageSizeSelector={false}
                            totalCount={importableSettings.length}
                        />
                    </div>
                )}

                <DialogFooter>
                    <Button
                        type="button"
                        variant="outline"
                        disabled={isPending}
                        onClick={handleClose}
                    >
                        Cancel
                    </Button>
                    <Button
                        type="button"
                        disabled={selectedIds.size === 0 || hasError}
                        isLoading={isPending}
                        onClick={handleImport}
                    >
                        Import
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
