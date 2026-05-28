import React, { type PropsWithChildren, useEffect, useImperativeHandle, useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { NotificationQueries } from "~/settings/data";
import type { SystemCleanupSettings } from "~/system-settings/domain";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import {
    type SystemCleanupConfigurationFormInput,
    type SystemCleanupConfigurationFormOutput,
    SystemCleanupConfigurationFormSchema,
} from "../schemas";
import type { SystemCleanupConfigurationFormRef } from "../types";

import {
    emptySystemCleanupConfigurationFormDefaults,
    mapSystemCleanupSettingsToFormInput,
} from "./system-cleanup-configuration.form-mappers";

type SchemaInput = SystemCleanupConfigurationFormInput;
type SchemaOutput = SystemCleanupConfigurationFormOutput;

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
    const {
        field: scheduleInterval,
        fieldState: { error: scheduleIntervalError, invalid: isScheduleIntervalInvalid },
    } = useController({ control, name: "scheduleInterval" });
    const {
        field: scheduleFrom,
        fieldState: { error: scheduleFromError, invalid: isScheduleFromInvalid },
    } = useController({ control, name: "scheduleFrom" });

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
                                    value={scheduleFrom.value ?? undefined}
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
            </div>
        </>
    );
}

function DBCleanupOptionsFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: enabled } = useController({ control, name: "dbObjectRetention.enabled" });
    const {
        field: tasks,
        fieldState: { error: tasksError, invalid: isTasksInvalid },
    } = useController({ control, name: "dbObjectRetention.tasks" });
    const {
        field: deployments,
        fieldState: { error: deploymentsError, invalid: isDeploymentsInvalid },
    } = useController({ control, name: "dbObjectRetention.deployments" });
    const {
        field: sysErrors,
        fieldState: { error: sysErrorsError, invalid: isSysErrorsInvalid },
    } = useController({ control, name: "dbObjectRetention.sysErrors" });
    const {
        field: deletedObjects,
        fieldState: { error: deletedObjectsError, invalid: isDeletedObjectsInvalid },
    } = useController({ control, name: "dbObjectRetention.deletedObjects" });

    return (
        <>
            <SectionHeader>DB Cleanup Options</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Enabled">
                    <Checkbox
                        checked={enabled.value}
                        onCheckedChange={enabled.onChange}
                    />
                </InfoBlock>

                {enabled.value && (
                    <>
                        <InfoBlock title="Task Object Retention">
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...tasks}
                                        placeholder="180d"
                                        className="max-w-[400px]"
                                        aria-invalid={isTasksInvalid}
                                    />
                                    <FieldError errors={[tasksError]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock title="Deployment Object Retention">
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...deployments}
                                        placeholder="180d"
                                        className="max-w-[400px]"
                                        aria-invalid={isDeploymentsInvalid}
                                    />
                                    <FieldError errors={[deploymentsError]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock title="Sys Error Object Retention">
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...sysErrors}
                                        placeholder="180d"
                                        className="max-w-[400px]"
                                        aria-invalid={isSysErrorsInvalid}
                                    />
                                    <FieldError errors={[sysErrorsError]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock title="Deleted Object Retention">
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...deletedObjects}
                                        placeholder="180d"
                                        className="max-w-[400px]"
                                        aria-invalid={isDeletedObjectsInvalid}
                                    />
                                    <FieldError errors={[deletedObjectsError]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </>
                )}
            </div>
        </>
    );
}

function DockerSwarmCleanupOptionsFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: enabled } = useController({ control, name: "clusterCleanup.enabled" });
    const { field: pruneImages } = useController({ control, name: "clusterCleanup.pruneImages" });
    const { field: pruneVolumes } = useController({ control, name: "clusterCleanup.pruneVolumes" });
    const { field: pruneNetworks } = useController({ control, name: "clusterCleanup.pruneNetworks" });
    const { field: pruneContainers } = useController({ control, name: "clusterCleanup.pruneContainers" });

    return (
        <>
            <SectionHeader>Docker Swarm Cleanup Options</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Enabled">
                    <Checkbox
                        checked={enabled.value}
                        onCheckedChange={enabled.onChange}
                    />
                </InfoBlock>

                {enabled.value && (
                    <>
                        <InfoBlock title="Prune Images">
                            <Checkbox
                                checked={pruneImages.value}
                                onCheckedChange={pruneImages.onChange}
                            />
                        </InfoBlock>

                        <InfoBlock title="Prune Volumes">
                            <Checkbox
                                checked={pruneVolumes.value}
                                onCheckedChange={pruneVolumes.onChange}
                            />
                        </InfoBlock>

                        <InfoBlock title="Prune Networks">
                            <Checkbox
                                checked={pruneNetworks.value}
                                onCheckedChange={pruneNetworks.onChange}
                            />
                        </InfoBlock>

                        <InfoBlock title="Prune Containers">
                            <Checkbox
                                checked={pruneContainers.value}
                                onCheckedChange={pruneContainers.onChange}
                            />
                        </InfoBlock>
                    </>
                )}
            </div>
        </>
    );
}

function BackupCleanupOptionsFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: enabled } = useController({ control, name: "backupCleanup.enabled" });
    const {
        field: localBackupRetention,
        fieldState: { error: localBackupRetentionError, invalid: isLocalBackupRetentionInvalid },
    } = useController({ control, name: "backupCleanup.localBackupRetention" });
    const {
        field: cloudBackupRetention,
        fieldState: { error: cloudBackupRetentionError, invalid: isCloudBackupRetentionInvalid },
    } = useController({ control, name: "backupCleanup.cloudBackupRetention" });

    return (
        <>
            <SectionHeader>Backup Cleanup Options</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Enabled">
                    <Checkbox
                        checked={enabled.value}
                        onCheckedChange={enabled.onChange}
                    />
                </InfoBlock>

                {enabled.value && (
                    <>
                        <InfoBlock title="Local Backup Retention">
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...localBackupRetention}
                                        placeholder="30d"
                                        className="max-w-[400px]"
                                        aria-invalid={isLocalBackupRetentionInvalid}
                                    />
                                    <FieldError errors={[localBackupRetentionError]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock title="Cloud Backup Retention">
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...cloudBackupRetention}
                                        placeholder="30d"
                                        className="max-w-[400px]"
                                        aria-invalid={isCloudBackupRetentionInvalid}
                                    />
                                    <FieldError errors={[cloudBackupRetentionError]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </>
                )}
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

function EnabledCleanupConfigurationFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const status = useWatch({ control, name: "status" });

    if (status !== ESettingStatus.Active) {
        return null;
    }

    return (
        <>
            <GeneralFields />
            <DBCleanupOptionsFields />
            <DockerSwarmCleanupOptionsFields />
            <BackupCleanupOptionsFields />
            <NotificationFields />
        </>
    );
}

function useSystemCleanupFormMethods(defaultValues?: SystemCleanupSettings) {
    return useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSystemCleanupSettingsToFormInput(defaultValues)
            : emptySystemCleanupConfigurationFormDefaults,
        resolver: zodResolver(SystemCleanupConfigurationFormSchema),
        mode: "onSubmit",
    });
}

export function SystemCleanupConfigurationForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useSystemCleanupFormMethods(defaultValues);

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
                    <EnabledCleanupConfigurationFields />
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SystemCleanupConfigurationFormRef>;
    defaultValues?: SystemCleanupSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
