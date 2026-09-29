import React, { type PropsWithChildren, useId, useImperativeHandle, useMemo, useState } from "react";

import { PasswordInput, RevealSecretsProvider } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { BackupRepoQueries } from "~/settings/data";
import { ConfirmRevealSecretsDialog, RevealSecretsButton } from "~/settings/module-shared/components";
import { useNotificationSettingsSources, useSettingRevealSecrets } from "~/settings/module-shared/hooks";
import { SYSTEM_BACKUP_SPEC_SECRETS, type SystemBackupSettings } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { AppLink, Combobox, InfoBlock, NextRunsField } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { NotificationSettings } from "@application/shared/form";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Separator } from "@/components/ui/separator";

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

function EnabledField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: status } = useController({ control, name: "status" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Enabled"
        >
            <Checkbox
                checked={status.value === ESettingStatus.Active}
                onCheckedChange={checked => {
                    status.onChange(checked ? ESettingStatus.Active : ESettingStatus.Disabled);
                }}
            />
        </InfoBlock>
    );
}

function GeneralFields({ nextRuns }: { nextRuns: Date[] }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const [repositorySearch, setRepositorySearch] = useState("");
    const includeDBId = useId();
    const includeSpecId = useId();

    const {
        data: { data: repositories } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = BackupRepoQueries.useFindManyPaginated({ search: repositorySearch });

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
    const {
        field: includeDB,
        fieldState: { error: includeError },
    } = useController({ control, name: "includeDB" });
    const { field: includeSpec } = useController({ control, name: "includeSpec" });
    const { field: specSecrets } = useController({ control, name: "specSecrets" });
    const {
        field: specPassphrase,
        fieldState: { error: specPassphraseError, invalid: isSpecPassphraseInvalid },
    } = useController({ control, name: "specPassphrase" });
    const {
        field: targetRepository,
        fieldState: { error: targetRepositoryError, invalid: isTargetRepositoryInvalid },
    } = useController({ control, name: "targetRepository" });

    const { canShowRevealButton, isDialogOpen, setIsDialogOpen, isRevealing, isRevealed, handleConfirmReveal } =
        useSettingRevealSecrets<SystemBackupSettings>({
            customPath: "/system/settings/backup",
            mode: "edit",
            onSuccess: data => {
                if (data.specPassphrase) {
                    specPassphrase.onChange(data.specPassphrase);
                }
            },
        });

    const isEncrypted = includeSpec.value && specSecrets.value === SYSTEM_BACKUP_SPEC_SECRETS.Encrypted;

    const repositoryOptions = useMemo(() => {
        return repositories.map(item => ({
            value: { id: item.id, name: item.name },
            label: item.name,
        }));
    }, [repositories]);

    return (
        <>
            <SectionHeader>General</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-orange-500">Note:</span>{" "}
                    <span>
                        We encourage you to run this task during low server load periods (e.g., midnight). Additionally,
                        you should schedule system tasks at different times (e.g., system cleanup at 1 AM, followed by
                        data backup at 2 AM).
                    </span>
                </div>

                <InfoBlock
                    titleWidth={220}
                    title="Scheduling Mode"
                >
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
                    <InfoBlock
                        titleWidth={220}
                        title="Scheduling Interval"
                    >
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
                    <InfoBlock
                        titleWidth={220}
                        title="Cron Expression"
                    >
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

                <InfoBlock
                    titleWidth={220}
                    title="Schedule From"
                >
                    <FieldGroup>
                        <Field>
                            <DateTimePicker
                                value={scheduleFrom.value ?? undefined}
                                onChange={date => {
                                    scheduleFrom.onChange(date ?? null);
                                }}
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

                <NextRunsField
                    nextRuns={nextRuns}
                    titleWidth={220}
                />

                <Separator className="opacity-50" />

                <InfoBlock
                    titleWidth={220}
                    title="Back Up"
                >
                    <FieldGroup>
                        <Field>
                            <div className="flex flex-col gap-2 text-sm">
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        id={includeDBId}
                                        checked={includeDB.value}
                                        onCheckedChange={checked => {
                                            includeDB.onChange(checked === true);
                                        }}
                                    />
                                    <label htmlFor={includeDBId}>
                                        Database: a dump of HivePaaS&apos;s own database
                                    </label>
                                </div>
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        id={includeSpecId}
                                        checked={includeSpec.value}
                                        onCheckedChange={checked => {
                                            includeSpec.onChange(checked === true);
                                        }}
                                    />
                                    <label htmlFor={includeSpecId}>
                                        Spec: the configuration of every project, env and app, as Export writes it
                                    </label>
                                </div>
                            </div>
                            <FieldError errors={[includeError]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                {includeSpec.value && (
                    <InfoBlock
                        titleWidth={220}
                        title="Secrets in the Spec"
                    >
                        <div className="flex flex-col gap-2">
                            <Tabs
                                value={specSecrets.value}
                                onValueChange={specSecrets.onChange}
                            >
                                <TabsList>
                                    <TabsTrigger value={SYSTEM_BACKUP_SPEC_SECRETS.Encrypted}>Encrypted</TabsTrigger>
                                    <TabsTrigger value={SYSTEM_BACKUP_SPEC_SECRETS.Omit}>Omit</TabsTrigger>
                                    <TabsTrigger value={SYSTEM_BACKUP_SPEC_SECRETS.Plaintext}>Plaintext</TabsTrigger>
                                </TabsList>
                            </Tabs>
                            <p className="text-sm text-muted-foreground max-w-[600px]">
                                {specSecrets.value === SYSTEM_BACKUP_SPEC_SECRETS.Encrypted &&
                                    "The spec is encrypted with a passphrase of its own, besides the repository's password. A lost passphrase is a spec whose secrets cannot be read."}
                                {specSecrets.value === SYSTEM_BACKUP_SPEC_SECRETS.Omit &&
                                    "The spec holds no secret: importing it asks for every one again."}
                                {specSecrets.value === SYSTEM_BACKUP_SPEC_SECRETS.Plaintext &&
                                    "The spec holds every secret as it is: whoever has the repository's password reads them."}
                            </p>
                        </div>
                    </InfoBlock>
                )}

                {isEncrypted && (
                    <InfoBlock
                        titleWidth={220}
                        title="Spec Passphrase"
                    >
                        <FieldGroup>
                            <Field>
                                <div className="flex items-center gap-2">
                                    <div className="w-full max-w-[400px]">
                                        <RevealSecretsProvider value={{ isRevealed }}>
                                            <PasswordInput
                                                value={specPassphrase.value}
                                                onChange={specPassphrase.onChange}
                                                placeholder="passphrase"
                                                className="w-full"
                                                aria-invalid={isSpecPassphraseInvalid}
                                            />
                                        </RevealSecretsProvider>
                                    </div>
                                    {canShowRevealButton && (
                                        <RevealSecretsButton
                                            onClick={() => {
                                                setIsDialogOpen(true);
                                            }}
                                            isLoading={isRevealing}
                                        />
                                    )}
                                </div>
                                <FieldError errors={[specPassphraseError]} />
                            </Field>
                        </FieldGroup>
                        <ConfirmRevealSecretsDialog
                            open={isDialogOpen}
                            onOpenChange={setIsDialogOpen}
                            onConfirm={handleConfirmReveal}
                            isPending={isRevealing}
                        />
                    </InfoBlock>
                )}

                <InfoBlock
                    titleWidth={220}
                    title="Backup Repository"
                >
                    <FieldGroup>
                        <Field>
                            <Combobox
                                options={repositoryOptions}
                                value={targetRepository.value?.id ?? null}
                                onChange={(_, option) => {
                                    targetRepository.onChange(option ?? undefined);
                                }}
                                onSearch={setRepositorySearch}
                                placeholder={targetRepository.value?.name ?? "Select a backup repository"}
                                emptyText="No backup repositories at the global scope"
                                className="max-w-[400px]"
                                valueKey="id"
                                searchable
                                closeOnSelect
                                loading={isFetching}
                                onRefresh={() => void refetch()}
                                isRefreshing={isRefetching}
                                aria-invalid={isTargetRepositoryInvalid}
                            />
                            <FieldError errors={[targetRepositoryError]} />
                            <div className="flex gap-4">
                                <AppLink.Basic
                                    to={ROUTE.settings.backupRepos.$route}
                                    target="_blank"
                                    className="text-xs text-blue-500"
                                    ignorePrevPath
                                >
                                    Configure Backup Repos
                                </AppLink.Basic>
                                <AppLink.Basic
                                    to={ROUTE.appSettings.dataBackup.snapshots.$route}
                                    className="text-xs text-blue-500"
                                    ignorePrevPath
                                >
                                    View snapshots
                                </AppLink.Basic>
                            </div>
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
                    titleWidth={220}
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
