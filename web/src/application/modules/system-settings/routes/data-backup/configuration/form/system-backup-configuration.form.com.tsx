import React, { type PropsWithChildren, useEffect, useImperativeHandle, useMemo, useState } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { CloudStorageQueries, NotificationQueries } from "~/settings/data";
import type { SystemBackupSettings } from "~/system-settings/domain";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../../module-shared/enums";
import {
    type SystemBackupConfigurationFormInput,
    type SystemBackupConfigurationFormOutput,
    SystemBackupConfigurationFormSchema,
} from "../schemas";
import type { SystemBackupConfigurationFormRef } from "../types";

import {
    emptySystemBackupConfigurationFormDefaults,
    mapSystemBackupSettingsToFormInput,
} from "./system-backup-configuration.form-mappers";

type SchemaInput = SystemBackupConfigurationFormInput;
type SchemaOutput = SystemBackupConfigurationFormOutput;

function SectionHeader({ children }: PropsWithChildren) {
    return <div className="rounded-lg bg-muted px-4 py-3 text-sm font-semibold text-foreground">{children}</div>;
}

function EnabledField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: status } = useController({ control, name: "status" });

    return (
        <InfoBlock title="Enabled">
            <Checkbox
                checked={status.value === ESettingStatus.Active}
                onCheckedChange={checked => {
                    status.onChange(checked ? ESettingStatus.Active : ESettingStatus.Disabled);
                }}
            />
        </InfoBlock>
    );
}

function GeneralFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const [cloudStorageSearch, setCloudStorageSearch] = useState("");

    const {
        data: { data: cloudStorages } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = CloudStorageQueries.useFindManyPaginated({ search: cloudStorageSearch });

    const {
        field: scheduleInterval,
        fieldState: { error: scheduleIntervalError, invalid: isScheduleIntervalInvalid },
    } = useController({ control, name: "scheduleInterval" });
    const {
        field: scheduleFrom,
        fieldState: { error: scheduleFromError, invalid: isScheduleFromInvalid },
    } = useController({ control, name: "scheduleFrom" });
    const { field: compressionFormat } = useController({ control, name: "compressionFormat" });
    const { field: encryptionFormat } = useController({ control, name: "encryptionFormat" });
    const {
        field: encryptionSecret,
        fieldState: { error: encryptionSecretError, invalid: isEncryptionSecretInvalid },
    } = useController({ control, name: "encryptionSecret" });
    const {
        field: cloudStorage,
        fieldState: { error: cloudStorageError, invalid: isCloudStorageInvalid },
    } = useController({ control, name: "cloudStorage" });
    const {
        field: cloudStorageDestinationDir,
        fieldState: { error: cloudStorageDestinationDirError, invalid: isCloudStorageDestinationDirInvalid },
    } = useController({ control, name: "cloudStorageDestinationDir" });

    useEffect(() => {
        if (encryptionFormat.value !== ESystemBackupEncryptionFormat.Age) {
            encryptionSecret.onChange("");
        }
    }, [encryptionFormat.value, encryptionSecret]);

    const cloudStorageOptions = useMemo(() => {
        return cloudStorages.map(item => ({
            value: { id: item.id, name: item.name },
            label: item.name,
        }));
    }, [cloudStorages]);

    return (
        <>
            <SectionHeader>General</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Run Interval">
                    <FieldGroup>
                        <Field>
                            <Input
                                {...scheduleInterval}
                                placeholder="1d"
                                className="max-w-[400px]"
                                aria-invalid={isScheduleIntervalInvalid}
                            />
                            <FieldError errors={[scheduleIntervalError]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock title="First Run Starts At">
                    <FieldGroup>
                        <Field>
                            <div className="flex items-center gap-4">
                                <DateTimePicker
                                    value={scheduleFrom.value}
                                    onChange={scheduleFrom.onChange}
                                    placeholder="select date time"
                                    granularity="minute"
                                    showClearButton
                                    aria-invalid={isScheduleFromInvalid}
                                    containerClassName="max-w-[400px]"
                                />
                                <Button
                                    type="button"
                                    variant="link"
                                    className="px-0"
                                >
                                    See Next Runs
                                </Button>
                            </div>
                            <FieldError errors={[scheduleFromError]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <div className="border-t" />

                <InfoBlock title="Compress Backups">
                    <Tabs
                        value={compressionFormat.value}
                        onValueChange={compressionFormat.onChange}
                    >
                        <TabsList>
                            <TabsTrigger value={ESystemBackupCompressionFormat.None}>Disabled</TabsTrigger>
                            <TabsTrigger value={ESystemBackupCompressionFormat.Gzip}>Gzip</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                <InfoBlock title="Encrypt Backups">
                    <Tabs
                        value={encryptionFormat.value}
                        onValueChange={encryptionFormat.onChange}
                    >
                        <TabsList>
                            <TabsTrigger value={ESystemBackupEncryptionFormat.None}>Disabled</TabsTrigger>
                            <TabsTrigger value={ESystemBackupEncryptionFormat.Age}>Age</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {encryptionFormat.value !== ESystemBackupEncryptionFormat.None && (
                    <InfoBlock title="Encryption Secret">
                        <FieldGroup>
                            <Field className="max-w-[400px]">
                                <PasswordInput
                                    value={encryptionSecret.value}
                                    onChange={encryptionSecret.onChange}
                                    placeholder="password"
                                    className="max-w-[400px]"
                                    aria-invalid={isEncryptionSecretInvalid}
                                />
                                <FieldError errors={[encryptionSecretError]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                )}

                <div className="border-t" />

                <InfoBlock title="Save Backups in Cloud Storage">
                    <FieldGroup>
                        <Field>
                            <Combobox
                                options={cloudStorageOptions}
                                value={cloudStorage.value?.id ?? null}
                                onChange={(_, option) => {
                                    cloudStorage.onChange(option ?? undefined);
                                }}
                                onSearch={setCloudStorageSearch}
                                placeholder="Select Cloud Storage"
                                emptyText="No cloud storages available"
                                className="max-w-[400px]"
                                valueKey="id"
                                searchable
                                closeOnSelect
                                allowClear
                                loading={isFetching}
                                onRefresh={() => void refetch()}
                                isRefreshing={isRefetching}
                                aria-invalid={isCloudStorageInvalid}
                            />
                            <FieldError errors={[cloudStorageError]} />
                            <AppLink.Basic
                                to={ROUTE.settings.cloudStorages.$route}
                                className="text-sm text-blue-500"
                                ignorePrevPath
                            >
                                Manage Cloud Storage Settings
                            </AppLink.Basic>
                        </Field>
                        <Field>
                            <Input
                                {...cloudStorageDestinationDir}
                                placeholder="Destination directory"
                                className="max-w-[400px]"
                                aria-invalid={isCloudStorageDestinationDirInvalid}
                            />
                            <FieldError errors={[cloudStorageDestinationDirError]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>
            </div>
        </>
    );
}

function EnabledBackupConfigurationFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const status = useWatch({ control, name: "status" });

    if (status !== ESettingStatus.Active) {
        return null;
    }

    return (
        <>
            <GeneralFields />
            <BackupOptionsFields />
            <NotificationFields />
        </>
    );
}

function BackupOptionsFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: backupDeletedObjects } = useController({ control, name: "backupDeletedObjects" });

    return (
        <>
            <SectionHeader>Backup Options</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Backup Deleted DB Records">
                    <Checkbox
                        checked={backupDeletedObjects.value}
                        onCheckedChange={backupDeletedObjects.onChange}
                    />
                </InfoBlock>
            </div>
        </>
    );
}

function NotificationFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const [successSearch, setSuccessSearch] = useState("");
    const [failureSearch, setFailureSearch] = useState("");

    const successQuery = NotificationQueries.useFindManyPaginated({ search: successSearch });
    const failureQuery = NotificationQueries.useFindManyPaginated({ search: failureSearch });

    const { field: successUseDefault } = useController({ control, name: "notification.successUseDefault" });
    const { field: success } = useController({ control, name: "notification.success" });
    const { field: failureUseDefault } = useController({ control, name: "notification.failureUseDefault" });
    const { field: failure } = useController({ control, name: "notification.failure" });

    useEffect(() => {
        if (successUseDefault.value) {
            success.onChange(undefined);
        }
    }, [successUseDefault.value, success]);

    useEffect(() => {
        if (failureUseDefault.value) {
            failure.onChange(undefined);
        }
    }, [failureUseDefault.value, failure]);

    const successOptions = useMemo(() => {
        return (successQuery.data ?? DEFAULT_PAGINATED_DATA).data.map(item => ({
            value: { id: item.id, name: item.name },
            label: item.name,
        }));
    }, [successQuery.data]);

    const failureOptions = useMemo(() => {
        return (failureQuery.data ?? DEFAULT_PAGINATED_DATA).data.map(item => ({
            value: { id: item.id, name: item.name },
            label: item.name,
        }));
    }, [failureQuery.data]);

    return (
        <>
            <SectionHeader>Notification Configuration</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="On Success Use Default"
                            content="Use the default notification settings on success"
                        />
                    }
                >
                    <Checkbox
                        checked={successUseDefault.value}
                        onCheckedChange={successUseDefault.onChange}
                    />
                </InfoBlock>

                <InfoBlock title="On Success">
                    <Field>
                        <Combobox
                            options={successOptions}
                            value={success.value?.id ?? null}
                            onChange={(_, option) => {
                                success.onChange(option ?? undefined);
                            }}
                            onSearch={setSuccessSearch}
                            placeholder="None"
                            emptyText="No notifications available"
                            className="max-w-[400px]"
                            valueKey="id"
                            searchable
                            closeOnSelect
                            allowClear
                            loading={successQuery.isFetching}
                            onRefresh={() => void successQuery.refetch()}
                            isRefreshing={successQuery.isRefetching}
                            disabled={successUseDefault.value}
                        />
                        <AppLink.Basic
                            to={ROUTE.settings.notificationTargets.$route}
                            className="text-sm text-blue-500"
                            ignorePrevPath
                        >
                            Manage Notification Settings
                        </AppLink.Basic>
                    </Field>
                </InfoBlock>

                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="On Failure Use Default"
                            content="Use the default notification settings on failure"
                        />
                    }
                >
                    <Checkbox
                        checked={failureUseDefault.value}
                        onCheckedChange={failureUseDefault.onChange}
                    />
                </InfoBlock>

                <InfoBlock title="On Failure">
                    <Field>
                        <Combobox
                            options={failureOptions}
                            value={failure.value?.id ?? null}
                            onChange={(_, option) => {
                                failure.onChange(option ?? undefined);
                            }}
                            onSearch={setFailureSearch}
                            placeholder="None"
                            emptyText="No notifications available"
                            className="max-w-[400px]"
                            valueKey="id"
                            searchable
                            closeOnSelect
                            allowClear
                            loading={failureQuery.isFetching}
                            onRefresh={() => void failureQuery.refetch()}
                            isRefreshing={failureQuery.isRefetching}
                            disabled={failureUseDefault.value}
                        />
                        <AppLink.Basic
                            to={ROUTE.settings.notificationTargets.$route}
                            className="text-sm text-blue-500"
                            ignorePrevPath
                        >
                            Manage Notification Settings
                        </AppLink.Basic>
                    </Field>
                </InfoBlock>
            </div>
        </>
    );
}

function useSystemBackupFormMethods(defaultValues?: SystemBackupSettings) {
    return useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSystemBackupSettingsToFormInput(defaultValues)
            : emptySystemBackupConfigurationFormDefaults,
        resolver: zodResolver(SystemBackupConfigurationFormSchema),
        mode: "onSubmit",
    });
}

export function SystemBackupConfigurationForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useSystemBackupFormMethods(defaultValues);

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<SchemaInput>) => {
                methods.reset({
                    ...methods.getValues(),
                    ...values,
                } as SchemaInput);
            },
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }

                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as FieldPath<SchemaInput>,
                        { message, type: "manual" },
                        {
                            shouldFocus: index === 0,
                        },
                    );
                });
            },
        }),
        [methods],
    );

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <EnabledField />
                    <EnabledBackupConfigurationFields />
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SystemBackupConfigurationFormRef>;
    defaultValues?: SystemBackupSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
