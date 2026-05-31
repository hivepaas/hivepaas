import React, { type PropsWithChildren, useEffect, useImperativeHandle, useMemo, useState } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { ChevronDown } from "lucide-react";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { CloudStorageQueries } from "~/settings/data";
import { useNotificationSettingsSources } from "~/settings/module-shared/hooks";
import type { SystemBackupSettings } from "~/system-settings/domain";

import { AppLink, Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { NotificationSettings } from "@application/shared/form";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    Checkbox,
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
    Field,
    FieldError,
    FieldGroup,
    Input,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../../module-shared/enums";
import {
    type SystemBackupConfigurationFormInput,
    type SystemBackupConfigurationFormOutput,
    SystemBackupConfigurationFormSchema,
    SystemBackupScheduleMode,
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

function NextRunsField({ nextRuns }: { nextRuns: Date[] }) {
    const [open, setOpen] = useState(true);

    return (
        <InfoBlock title="Next Runs">
            <Collapsible
                open={open}
                onOpenChange={setOpen}
                className="max-w-[400px]"
            >
                <div className={cn(dashedBorderBox, "text-center text-sm leading-6 p-2")}>
                    <CollapsibleTrigger asChild>
                        <button
                            type="button"
                            aria-label={open ? "Collapse next runs" : "Expand next runs"}
                            className="flex w-full items-center justify-end"
                        >
                            <ChevronDown
                                className={cn(
                                    "size-4 text-muted-foreground transition-transform duration-200",
                                    open && "rotate-180",
                                )}
                            />
                        </button>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                        {nextRuns.length > 0 ? (
                            <div className="flex flex-col gap-1 pt-3 text-sm">
                                {nextRuns.map(runAt => (
                                    <span
                                        key={runAt.toISOString()}
                                        className="text-orange-500"
                                    >
                                        {format(runAt, "yyyy-MM-dd HH:mm:ss")}
                                    </span>
                                ))}
                            </div>
                        ) : (
                            <div className="pt-3 text-sm text-muted-foreground">No next runs available</div>
                        )}
                    </CollapsibleContent>
                </div>
            </Collapsible>
        </InfoBlock>
    );
}

function GeneralFields({ nextRuns }: { nextRuns: Date[] }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const [cloudStorageSearch, setCloudStorageSearch] = useState("");

    const {
        data: { data: cloudStorages } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = CloudStorageQueries.useFindManyPaginated({ search: cloudStorageSearch });

    const { field: scheduleMode } = useController({ control, name: "scheduleMode" });
    const {
        field: scheduleInterval,
        fieldState: { error: scheduleIntervalError, invalid: isScheduleIntervalInvalid },
    } = useController({ control, name: "scheduleInterval" });
    const {
        field: scheduleCronExpr,
        fieldState: { error: scheduleCronExprError, invalid: isScheduleCronExprInvalid },
    } = useController({ control, name: "scheduleCronExpr" });
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
                <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                    <span className="text-orange-500">Note:</span>{" "}
                    <span>
                        We encourage you to run this task during low server load periods (e.g., midnight). Additionally,
                        you should schedule system tasks at different times (e.g., system cleanup at 1 AM, followed by
                        data backup at 2 AM).
                    </span>
                </div>

                <InfoBlock title="Scheduling Mode">
                    <Tabs
                        value={scheduleMode.value}
                        onValueChange={scheduleMode.onChange}
                    >
                        <TabsList>
                            <TabsTrigger value={SystemBackupScheduleMode.Interval}>Interval-based</TabsTrigger>
                            <TabsTrigger value={SystemBackupScheduleMode.Cron}>Time-based</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {scheduleMode.value === SystemBackupScheduleMode.Interval && (
                    <InfoBlock title="Scheduling Interval">
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...scheduleInterval}
                                    placeholder="1d, 1h30m"
                                    className="max-w-[400px]"
                                    aria-invalid={isScheduleIntervalInvalid}
                                />
                                <FieldError errors={[scheduleIntervalError]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                )}

                {scheduleMode.value === SystemBackupScheduleMode.Cron && (
                    <InfoBlock title="Cron Expression">
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...scheduleCronExpr}
                                    placeholder="accepted form: * * * * *"
                                    className="max-w-[400px]"
                                    aria-invalid={isScheduleCronExprInvalid}
                                />
                                <FieldError errors={[scheduleCronExprError]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                )}

                <InfoBlock title="Schedule From">
                    <FieldGroup>
                        <Field>
                            <DateTimePicker
                                value={scheduleFrom.value ?? undefined}
                                onChange={scheduleFrom.onChange}
                                placeholder="select date time"
                                granularity="minute"
                                showClearButton
                                aria-invalid={isScheduleFromInvalid}
                                containerClassName="max-w-[400px]"
                            />
                            <FieldError errors={[scheduleFromError]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <NextRunsField nextRuns={nextRuns} />

                <div className="border-t" />

                <InfoBlock title="Compress Backups">
                    <Tabs
                        value={compressionFormat.value}
                        onValueChange={compressionFormat.onChange}
                    >
                        <TabsList>
                            <TabsTrigger value={ESystemBackupCompressionFormat.None}>Disabled</TabsTrigger>
                            <TabsTrigger value={ESystemBackupCompressionFormat.Gzip}>Gzip</TabsTrigger>
                            <TabsTrigger value={ESystemBackupCompressionFormat.Zstd}>Zstd</TabsTrigger>
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
                                placeholder="path/to/sub/dir"
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

function EnabledBackupConfigurationFields({ nextRuns, readOnly }: { nextRuns: Date[]; readOnly: boolean }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const status = useWatch({ control, name: "status" });

    if (status !== ESettingStatus.Active) {
        return null;
    }

    return (
        <>
            <GeneralFields nextRuns={nextRuns} />
            <NotificationFields readOnly={readOnly} />
        </>
    );
}

function NotificationFields({ readOnly = false }: { readOnly?: boolean }) {
    const { sources, manageLink } = useNotificationSettingsSources({ type: "settings" });

    return (
        <>
            <SectionHeader>Notification Configuration</SectionHeader>
            <div className="px-3">
                <NotificationSettings<SchemaInput>
                    names={{
                        successUseDefault: "notification.successUseDefault",
                        success: "notification.success",
                        failureUseDefault: "notification.failureUseDefault",
                        failure: "notification.failure",
                    }}
                    sources={sources}
                    manageLink={manageLink}
                    readOnly={readOnly}
                />
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

export function SystemBackupConfigurationForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
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
                        if (readOnly) {
                            return;
                        }

                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
                    >
                        <EnabledField />
                        <EnabledBackupConfigurationFields
                            nextRuns={defaultValues?.nextRuns ?? []}
                            readOnly={readOnly}
                        />
                    </fieldset>
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
    readOnly?: boolean;
}>;
