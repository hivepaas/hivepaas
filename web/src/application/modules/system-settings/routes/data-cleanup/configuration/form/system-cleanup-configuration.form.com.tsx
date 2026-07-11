import React, { type PropsWithChildren, useImperativeHandle, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { ChevronDown } from "lucide-react";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { useNotificationSettingsSources } from "~/settings/module-shared/hooks";
import type { SystemCleanupSettings } from "~/system-settings/domain";

import { InfoBlock } from "@application/shared/components";
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

import {
    type SystemCleanupConfigurationFormInput,
    type SystemCleanupConfigurationFormOutput,
    SystemCleanupConfigurationFormSchema,
    SystemCleanupScheduleMode,
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
                            <TabsTrigger value={SystemCleanupScheduleMode.Interval}>Interval-based</TabsTrigger>
                            <TabsTrigger value={SystemCleanupScheduleMode.Cron}>Time-based</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {scheduleMode.value === SystemCleanupScheduleMode.Interval && (
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

                {scheduleMode.value === SystemCleanupScheduleMode.Cron && (
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

                <NextRunsField nextRuns={nextRuns} />
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

function CacheCleanupOptionsFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: enabled } = useController({ control, name: "cacheCleanup.enabled" });
    const {
        field: repoCacheRetention,
        fieldState: { error: repoCacheRetentionError, invalid: isRepoCacheRetentionInvalid },
    } = useController({ control, name: "cacheCleanup.repoCacheRetention" });

    return (
        <>
            <SectionHeader>Cache Cleanup Options</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Enabled">
                    <Checkbox
                        checked={enabled.value}
                        onCheckedChange={enabled.onChange}
                    />
                </InfoBlock>

                {enabled.value && (
                    <InfoBlock title="Repo Cache Retention">
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...repoCacheRetention}
                                    placeholder="30d"
                                    className="max-w-[400px]"
                                    aria-invalid={isRepoCacheRetentionInvalid}
                                />
                                <FieldError errors={[repoCacheRetentionError]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                )}
            </div>
        </>
    );
}

function FileCleanupOptionsFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: enabled } = useController({ control, name: "fileCleanup.enabled" });

    return (
        <>
            <SectionHeader>File Cleanup Options</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock title="Enabled">
                    <Checkbox
                        checked={enabled.value}
                        onCheckedChange={enabled.onChange}
                    />
                </InfoBlock>
            </div>
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

function EnabledCleanupConfigurationFields({ nextRuns, readOnly }: { nextRuns: Date[]; readOnly: boolean }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const status = useWatch({ control, name: "status" });

    if (status !== ESettingStatus.Active) {
        return null;
    }

    return (
        <>
            <GeneralFields nextRuns={nextRuns} />
            <DBCleanupOptionsFields />
            <DockerSwarmCleanupOptionsFields />
            <BackupCleanupOptionsFields />
            <CacheCleanupOptionsFields />
            <FileCleanupOptionsFields />
            <NotificationFields readOnly={readOnly} />
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

export function SystemCleanupConfigurationForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
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
                        <EnabledCleanupConfigurationFields
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
    ref?: React.Ref<SystemCleanupConfigurationFormRef>;
    defaultValues?: SystemCleanupSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
