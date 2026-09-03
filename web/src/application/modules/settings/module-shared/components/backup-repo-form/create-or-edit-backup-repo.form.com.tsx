import { useEffect, useMemo } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import { ProjectCloudStorageQueries, ProjectClusterVolumesQueries } from "~/projects/data/queries";
import { CloudStorageQueries } from "~/settings/data/queries";
import { useUpdateBackupRepoPasswordDialog } from "~/settings/dialogs";
import {
    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
    SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS,
} from "~/settings/module-shared/constants/settings-form-layout.constants";

import {
    AppLink,
    Combobox,
    ContentBlock,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import {
    Button,
    Checkbox,
    Field,
    FieldError,
    FieldGroup,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";

import {
    BACKUP_REPO_ACTION,
    BACKUP_REPO_ENGINE,
    BACKUP_REPO_ENGINE_COMPRESSION_OPTIONS,
    BACKUP_REPO_ENGINE_OPTIONS,
    BACKUP_REPO_STORAGE_TYPE,
    DEFAULT_BACKUP_REPO_COMPRESSION,
    DEFAULT_BACKUP_REPO_PACK_SIZE,
    DEFAULT_BACKUP_REPO_RETENTION,
    KOPIA_COMPRESSION_LEVELS,
} from "../../constants/backup-repo.constants";
import type { BackupRepoTableScope } from "../backup-repo-table/backup-repo-table.types";
import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditBackupRepoFormInput,
    CreateOrEditBackupRepoFormOutput,
} from "./create-or-edit-backup-repo.form.schema";
import { CreateOrEditBackupRepoFormSchema } from "./create-or-edit-backup-repo.form.schema";

type StorageOption = { id: string; name?: string };

const PAGINATION_ALL = { page: 1, size: 10000 };

export function CreateOrEditBackupRepoForm({
    backupRepoId,
    mode,
    scope,
    initialValues,
    savedVersion,
    isPending,
    readOnly = false,
    readOnlyInherited = false,
    showAvailableInProjects,
    onSubmit,
    onHasChanges,
    onClose,
}: Props) {
    const updatePasswordDialog = useUpdateBackupRepoPasswordDialog();
    const isUpdate = mode === "edit";
    const isReadOnly = readOnly || readOnlyInherited;
    const projectId = scope.type === "project" ? scope.projectId : "";
    const env = scope.type === "project" ? scope.env : undefined;

    const methods = useForm<CreateOrEditBackupRepoFormInput, unknown, CreateOrEditBackupRepoFormOutput>({
        resolver: zodResolver(CreateOrEditBackupRepoFormSchema),
        defaultValues: {
            name: initialValues?.name ?? "",
            engine: initialValues?.engine ?? BACKUP_REPO_ENGINE.Kopia,
            action: initialValues?.action ?? BACKUP_REPO_ACTION.CreateNew,
            description: initialValues?.description ?? "",
            storageType: initialValues?.storageType ?? BACKUP_REPO_STORAGE_TYPE.CloudStorage,
            cloudStorage: initialValues?.cloudStorage ?? null,
            volume: initialValues?.volume ?? null,
            storagePrefix: initialValues?.storagePrefix ?? "",
            password: initialValues?.password ?? "",
            compression: initialValues?.compression ?? DEFAULT_BACKUP_REPO_COMPRESSION,
            packSize: initialValues?.packSize ?? DEFAULT_BACKUP_REPO_PACK_SIZE,
            retention: {
                keepLast: initialValues?.retention?.keepLast ?? DEFAULT_BACKUP_REPO_RETENTION.keepLast,
                keepHourly: initialValues?.retention?.keepHourly ?? DEFAULT_BACKUP_REPO_RETENTION.keepHourly,
                keepDaily: initialValues?.retention?.keepDaily ?? DEFAULT_BACKUP_REPO_RETENTION.keepDaily,
                keepWeekly: initialValues?.retention?.keepWeekly ?? DEFAULT_BACKUP_REPO_RETENTION.keepWeekly,
                keepMonthly: initialValues?.retention?.keepMonthly ?? DEFAULT_BACKUP_REPO_RETENTION.keepMonthly,
            },
            inheritable: initialValues?.inheritable ?? false,
            default: initialValues?.default ?? false,
        },
    });

    const {
        control,
        handleSubmit,
        reset,
        getValues,
        formState: { isDirty, errors },
    } = methods;

    const initialName = initialValues?.name;
    const initialEngine = initialValues?.engine;
    const initialAction = initialValues?.action;
    const initialDescription = initialValues?.description;
    const initialStorageType = initialValues?.storageType;
    const initialCloudStorage = initialValues?.cloudStorage;
    const initialVolume = initialValues?.volume;
    const initialStoragePrefix = initialValues?.storagePrefix;
    const initialPassword = initialValues?.password;
    const initialCompression = initialValues?.compression;
    const initialPackSize = initialValues?.packSize;
    const initialKeepLast = initialValues?.retention?.keepLast;
    const initialKeepHourly = initialValues?.retention?.keepHourly;
    const initialKeepDaily = initialValues?.retention?.keepDaily;
    const initialKeepWeekly = initialValues?.retention?.keepWeekly;
    const initialKeepMonthly = initialValues?.retention?.keepMonthly;
    const initialInheritable = initialValues?.inheritable;
    const initialDefault = initialValues?.default;

    useEffect(() => {
        if (savedVersion === 0) {
            return;
        }

        reset(getValues());
        onHasChanges?.(false);
    }, [getValues, onHasChanges, reset, savedVersion]);

    useEffect(() => {
        if (initialName === undefined || isDirty) {
            return;
        }

        reset({
            name: initialName,
            engine: initialEngine ?? BACKUP_REPO_ENGINE.Kopia,
            action: initialAction ?? BACKUP_REPO_ACTION.CreateNew,
            description: initialDescription ?? "",
            storageType: initialStorageType ?? BACKUP_REPO_STORAGE_TYPE.CloudStorage,
            cloudStorage: initialCloudStorage ?? null,
            volume: initialVolume ?? null,
            storagePrefix: initialStoragePrefix ?? "",
            password: initialPassword ?? "",
            compression: initialCompression ?? DEFAULT_BACKUP_REPO_COMPRESSION,
            packSize: initialPackSize ?? DEFAULT_BACKUP_REPO_PACK_SIZE,
            retention: {
                keepLast: initialKeepLast ?? DEFAULT_BACKUP_REPO_RETENTION.keepLast,
                keepHourly: initialKeepHourly ?? DEFAULT_BACKUP_REPO_RETENTION.keepHourly,
                keepDaily: initialKeepDaily ?? DEFAULT_BACKUP_REPO_RETENTION.keepDaily,
                keepWeekly: initialKeepWeekly ?? DEFAULT_BACKUP_REPO_RETENTION.keepWeekly,
                keepMonthly: initialKeepMonthly ?? DEFAULT_BACKUP_REPO_RETENTION.keepMonthly,
            },
            inheritable: initialInheritable ?? false,
            default: initialDefault ?? false,
        });
    }, [
        initialAction,
        initialCloudStorage,
        initialCompression,
        initialDefault,
        initialDescription,
        initialEngine,
        initialInheritable,
        initialKeepDaily,
        initialKeepHourly,
        initialKeepLast,
        initialKeepMonthly,
        initialKeepWeekly,
        initialName,
        initialPackSize,
        initialPassword,
        initialStoragePrefix,
        initialStorageType,
        initialVolume,
        isDirty,
        reset,
    ]);

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, isReadOnly, onHasChanges]);

    // Query Cloud Storages
    const globalCloudStorageQuery = CloudStorageQueries.useFindManyPaginated(
        { pagination: PAGINATION_ALL },
        { enabled: scope.type === "settings" },
    );
    const projectCloudStorageQuery = ProjectCloudStorageQueries.useFindManyPaginated(
        { projectID: projectId, env, pagination: PAGINATION_ALL },
        { enabled: scope.type === "project" },
    );
    const cloudStorageQuery = scope.type === "project" ? projectCloudStorageQuery : globalCloudStorageQuery;
    const cloudStorageOptions = useMemo(() => cloudStorageQuery.data?.data ?? [], [cloudStorageQuery.data?.data]);

    const cloudStorageComboboxOptions = useMemo(() => {
        return cloudStorageOptions.map(item => ({
            value: { id: item.id, name: item.name } satisfies StorageOption,
            label: item.name,
        }));
    }, [cloudStorageOptions]);

    // Query Volumes
    const globalVolumesQuery = ClusterVolumesQueries.useFindManyPaginated(
        { pagination: PAGINATION_ALL },
        { enabled: scope.type === "settings" },
    );
    const projectVolumesQuery = ProjectClusterVolumesQueries.useFindManyPaginated(
        { projectID: projectId, env, pagination: PAGINATION_ALL },
        { enabled: scope.type === "project" },
    );
    const volumeQuery = scope.type === "project" ? projectVolumesQuery : globalVolumesQuery;
    const volumeOptions = useMemo(() => volumeQuery.data?.data ?? [], [volumeQuery.data?.data]);

    const volumeComboboxOptions = useMemo(() => {
        return volumeOptions.map(item => ({
            value: { id: item.id, name: item.name } satisfies StorageOption,
            label: item.name,
        }));
    }, [volumeOptions]);

    // Route links for configure
    const cloudStorageManageRoute =
        scope.type === "project"
            ? ROUTE.projects.single.providerConfiguration.cloudStorages.$route(scope.projectId)
            : ROUTE.settings.cloudStorages.$route;

    const volumeManageRoute =
        scope.type === "project"
            ? ROUTE.projects.single.clusterResources.volumes.$route(scope.projectId)
            : ROUTE.cluster.volumes.$route;

    // Field controllers
    const {
        field: nameField,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const { field: engineField } = useController({ name: "engine", control });
    const { field: actionField } = useController({ name: "action", control });
    const { field: descriptionField } = useController({ name: "description", control });
    const { field: storageTypeField } = useController({ name: "storageType", control });
    const { field: cloudStorageField } = useController({ name: "cloudStorage", control });
    const { field: volumeField } = useController({ name: "volume", control });
    const { field: storagePrefixField } = useController({ name: "storagePrefix", control });
    const {
        field: passwordField,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({ name: "password", control });
    const { field: compressionField } = useController({ name: "compression", control });
    const { field: packSizeField } = useController({ name: "packSize", control });
    const { field: keepLastField } = useController({ name: "retention.keepLast", control });
    const { field: keepHourlyField } = useController({ name: "retention.keepHourly", control });
    const { field: keepDailyField } = useController({ name: "retention.keepDaily", control });
    const { field: keepWeeklyField } = useController({ name: "retention.keepWeekly", control });
    const { field: keepMonthlyField } = useController({ name: "retention.keepMonthly", control });
    const { field: inheritableField } = useController({ name: "inheritable", control });
    const { field: defaultField } = useController({ name: "default", control });

    const watchedAction = useWatch({ control, name: "action" });
    const watchedStorageType = useWatch({ control, name: "storageType" });
    const watchedEngine = useWatch({ control, name: "engine" });

    const isImportExisting = mode === "create" && watchedAction === BACKUP_REPO_ACTION.ImportExisting;

    const compressionOptions = useMemo(() => {
        const levels =
            (watchedEngine ? BACKUP_REPO_ENGINE_COMPRESSION_OPTIONS[watchedEngine] : undefined) ??
            KOPIA_COMPRESSION_LEVELS;
        return levels.map((level: string) => ({
            label: level,
            value: level,
        }));
    }, [watchedEngine]);

    function onValid(values: CreateOrEditBackupRepoFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditBackupRepoFormOutput>) {
        console.error(_errors);
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <div>
                    {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                    {readOnly && !readOnlyInherited && <PermissionReadonlyNotice />}

                    <fieldset
                        disabled={isReadOnly}
                        className={`flex flex-col gap-6 border-0 p-0 m-0 min-w-0 ${SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS}`}
                    >
                        {/* Top Block: Name, Engine, Action, Description */}
                        <FieldGroup>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Name"
                                        isRequired
                                    />
                                }
                            >
                                <Field data-invalid={isNameInvalid}>
                                    <Input
                                        {...nameField}
                                        placeholder="my backup repo"
                                        aria-invalid={isNameInvalid}
                                    />
                                    <FieldError errors={[errors.name]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Engine Type"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Select
                                        value={engineField.value}
                                        onValueChange={engineField.onChange}
                                        disabled
                                    >
                                        <SelectTrigger className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                            <SelectValue placeholder="Select engine" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {BACKUP_REPO_ENGINE_OPTIONS.map(
                                                (option: { label: string; value: string }) => (
                                                    <SelectItem
                                                        key={option.value}
                                                        value={option.value}
                                                    >
                                                        {option.label}
                                                    </SelectItem>
                                                ),
                                            )}
                                        </SelectContent>
                                    </Select>
                                </Field>
                            </InfoBlock>

                            {/* 2. Action: visible in Create view, hidden in Update view */}
                            {!isUpdate && (
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Action" />}
                                >
                                    <Field>
                                        <Tabs
                                            value={actionField.value}
                                            onValueChange={actionField.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger value={BACKUP_REPO_ACTION.CreateNew}>
                                                    Create New
                                                </TabsTrigger>
                                                <TabsTrigger value={BACKUP_REPO_ACTION.ImportExisting}>
                                                    Import Existing
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                    </Field>
                                </InfoBlock>
                            )}

                            {/* Description: hidden in Import Existing */}
                            {!isImportExisting && (
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Description" />}
                                >
                                    <Field>
                                        <Input
                                            {...descriptionField}
                                            placeholder="e.g. database backup, image folder backup"
                                        />
                                    </Field>
                                </InfoBlock>
                            )}
                        </FieldGroup>

                        {/* Section: Storage */}
                        <ContentBlock label="Storage">
                            <FieldGroup>
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Type" />}
                                >
                                    <Field>
                                        <Tabs
                                            value={storageTypeField.value}
                                            onValueChange={val => {
                                                if (!isUpdate) {
                                                    storageTypeField.onChange(val);
                                                }
                                            }}
                                        >
                                            <TabsList>
                                                <TabsTrigger
                                                    value={BACKUP_REPO_STORAGE_TYPE.CloudStorage}
                                                    disabled={isUpdate}
                                                >
                                                    Cloud Storage
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={BACKUP_REPO_STORAGE_TYPE.Volume}
                                                    disabled={isUpdate}
                                                >
                                                    Volume
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                    </Field>
                                </InfoBlock>

                                {watchedStorageType === BACKUP_REPO_STORAGE_TYPE.CloudStorage && (
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Cloud Storage" />}
                                    >
                                        <Field>
                                            <Combobox<StorageOption>
                                                options={cloudStorageComboboxOptions}
                                                value={cloudStorageField.value?.id ?? null}
                                                onChange={(_, option) => {
                                                    if (!isUpdate) {
                                                        cloudStorageField.onChange(option ?? null);
                                                    }
                                                }}
                                                placeholder="select cloud storage"
                                                searchable
                                                closeOnSelect
                                                emptyText="No cloud storages available"
                                                className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                valueKey="id"
                                                loading={cloudStorageQuery.isFetching}
                                                onRefresh={() => void cloudStorageQuery.refetch()}
                                                isRefreshing={cloudStorageQuery.isRefetching}
                                                disabled={isReadOnly || isUpdate}
                                            />
                                            <AppLink.Modules
                                                to={cloudStorageManageRoute}
                                                className="text-xs text-link"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                ignorePrevPath
                                            >
                                                Configure Cloud Storages
                                            </AppLink.Modules>
                                        </Field>
                                    </InfoBlock>
                                )}

                                {watchedStorageType === BACKUP_REPO_STORAGE_TYPE.Volume && (
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Volume" />}
                                    >
                                        <Field>
                                            <Combobox<StorageOption>
                                                options={volumeComboboxOptions}
                                                value={volumeField.value?.id ?? null}
                                                onChange={(_, option) => {
                                                    if (!isUpdate) {
                                                        volumeField.onChange(option ?? null);
                                                    }
                                                }}
                                                placeholder="select volume"
                                                searchable
                                                closeOnSelect
                                                emptyText="No volumes available"
                                                className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                valueKey="id"
                                                loading={volumeQuery.isFetching}
                                                onRefresh={() => void volumeQuery.refetch()}
                                                isRefreshing={volumeQuery.isRefetching}
                                                disabled={isReadOnly || isUpdate}
                                            />
                                            <AppLink.Modules
                                                to={volumeManageRoute}
                                                className="text-xs text-link"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                ignorePrevPath
                                            >
                                                Configure Volumes
                                            </AppLink.Modules>
                                        </Field>
                                    </InfoBlock>
                                )}

                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Storage Path" />}
                                >
                                    <Field>
                                        <Input
                                            {...storagePrefixField}
                                            placeholder="/path/to/sub/dir"
                                            disabled={isReadOnly || isUpdate}
                                        />
                                    </Field>
                                </InfoBlock>
                            </FieldGroup>
                        </ContentBlock>

                        {/* Section: Repo Options */}
                        <ContentBlock label="Repo Options">
                            <FieldGroup>
                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="Password"
                                            isRequired={!isUpdate}
                                        />
                                    }
                                >
                                    <Field data-invalid={isPasswordInvalid}>
                                        <div className="flex gap-2">
                                            <div className="flex-1 min-w-0">
                                                <PasswordInput
                                                    {...passwordField}
                                                    placeholder="********"
                                                    disabled={isReadOnly || isUpdate}
                                                    aria-invalid={isPasswordInvalid}
                                                />
                                            </div>
                                            {isUpdate && (
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    disabled={isReadOnly}
                                                    onClick={() => {
                                                        if (backupRepoId) {
                                                            updatePasswordDialog.actions.open(scope, backupRepoId);
                                                        }
                                                    }}
                                                    className="shrink-0"
                                                >
                                                    Change Password
                                                </Button>
                                            )}
                                        </div>
                                        <FieldError errors={[errors.password]} />
                                    </Field>
                                </InfoBlock>

                                {!isImportExisting && (
                                    <>
                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Compression Level" />}
                                        >
                                            <Field>
                                                <Select
                                                    value={compressionField.value}
                                                    onValueChange={compressionField.onChange}
                                                    disabled={isReadOnly}
                                                >
                                                    <SelectTrigger className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                                        <SelectValue placeholder="Select compression" />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        {compressionOptions.map(
                                                            (option: { label: string; value: string }) => (
                                                                <SelectItem
                                                                    key={option.value}
                                                                    value={option.value}
                                                                >
                                                                    {option.label}
                                                                </SelectItem>
                                                            ),
                                                        )}
                                                    </SelectContent>
                                                </Select>
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Pack Size" />}
                                        >
                                            <Field>
                                                <Input
                                                    {...packSizeField}
                                                    placeholder="32mb"
                                                    disabled={isReadOnly}
                                                />
                                            </Field>
                                        </InfoBlock>
                                    </>
                                )}
                            </FieldGroup>
                        </ContentBlock>

                        {/* Section: Retention Policy (hidden in Import Existing) */}
                        {!isImportExisting && (
                            <ContentBlock label="Retention Policy">
                                <div className="flex flex-col gap-4">
                                    <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                                        <span className="font-semibold text-orange-500">Note:</span> Retention rules
                                        apply together and independently — a backup is kept if it matches any enabled
                                        rule. Leaving a rule blank means no extra limit from that rule, not that data
                                        gets deleted.
                                    </div>

                                    <FieldGroup>
                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Keep Last" />}
                                        >
                                            <Field>
                                                <Input
                                                    {...keepLastField}
                                                    type="number"
                                                    min={0}
                                                    placeholder={String(DEFAULT_BACKUP_REPO_RETENTION.keepLast)}
                                                    disabled={isReadOnly}
                                                />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Keep Hourly" />}
                                        >
                                            <Field>
                                                <Input
                                                    {...keepHourlyField}
                                                    type="number"
                                                    min={0}
                                                    placeholder={String(DEFAULT_BACKUP_REPO_RETENTION.keepHourly)}
                                                    disabled={isReadOnly}
                                                />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Keep Daily" />}
                                        >
                                            <Field>
                                                <Input
                                                    {...keepDailyField}
                                                    type="number"
                                                    min={0}
                                                    placeholder={String(DEFAULT_BACKUP_REPO_RETENTION.keepDaily)}
                                                    disabled={isReadOnly}
                                                />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Keep Weekly" />}
                                        >
                                            <Field>
                                                <Input
                                                    {...keepWeeklyField}
                                                    type="number"
                                                    min={0}
                                                    placeholder={String(DEFAULT_BACKUP_REPO_RETENTION.keepWeekly)}
                                                    disabled={isReadOnly}
                                                />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Keep Monthly" />}
                                        >
                                            <Field>
                                                <Input
                                                    {...keepMonthlyField}
                                                    type="number"
                                                    min={0}
                                                    placeholder={String(DEFAULT_BACKUP_REPO_RETENTION.keepMonthly)}
                                                    disabled={isReadOnly}
                                                />
                                            </Field>
                                        </InfoBlock>
                                    </FieldGroup>
                                </div>
                            </ContentBlock>
                        )}

                        {/* Available In Projects & Default */}
                        <FieldGroup>
                            {showAvailableInProjects && (
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Available In Projects" />}
                                >
                                    <Checkbox
                                        checked={inheritableField.value}
                                        onCheckedChange={checked => {
                                            inheritableField.onChange(Boolean(checked));
                                        }}
                                        disabled={isReadOnly}
                                    />
                                </InfoBlock>
                            )}

                            <InfoBlock
                                titleWidth={220}
                                title={<LabelWithInfo label="Default" />}
                            >
                                <Checkbox
                                    checked={defaultField.value}
                                    onCheckedChange={checked => {
                                        defaultField.onChange(Boolean(checked));
                                    }}
                                    disabled={isReadOnly}
                                />
                            </InfoBlock>
                        </FieldGroup>
                    </fieldset>
                </div>

                {!isReadOnly && (
                    <FormActionBar>
                        <SettingsFormCancelAction
                            onCancel={onClose}
                            disabled={isPending}
                        />
                        <Button
                            type="submit"
                            isLoading={isPending}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </FormActionBar>
                )}

                {isReadOnly && (
                    <FormActionBar>
                        <Button
                            type="button"
                            onClick={onClose}
                            className="min-w-[100px]"
                        >
                            Close
                        </Button>
                    </FormActionBar>
                )}
            </form>
        </FormProvider>
    );
}

interface Props {
    backupRepoId?: string;
    mode: "create" | "edit";
    isPending: boolean;
    onSubmit: (values: CreateOrEditBackupRepoFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditBackupRepoFormInput>;
    scope: BackupRepoTableScope;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
