import React, { useId, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppScheduledJob } from "~/projects/domain";
import { EAppScheduledJobScheduleMode } from "~/projects/module-shared/enums";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList, NotificationSettings } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { InputNumber } from "@/components/ui/input-number";

import type { CreateOrEditAppScheduledJobFormInput, CreateOrEditAppScheduledJobFormOutput } from "../schemas";
import { APP_SCHEDULED_JOB_COMMAND_MODE, CreateOrEditAppScheduledJobFormSchema } from "../schemas";

import {
    ArgGroupsSection,
    INFO_BLOCK_TITLE_WIDTH,
    NextRunsField,
    PriorityTabsField,
    ScriptEditorField,
} from "./building-blocks";
import {
    createEmptyAppScheduledJobFormDefaults,
    mapAppScheduledJobToFormInput,
} from "./create-or-edit-app-scheduled-job.form-mappers";

type SchemaInput = CreateOrEditAppScheduledJobFormInput;
type SchemaOutput = CreateOrEditAppScheduledJobFormOutput;

export function CreateOrEditAppScheduledJobForm({
    projectId,
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    readOnly = false,
}: Props) {
    const defaultValues = useMemo(() => {
        if (initialValues) {
            return mapAppScheduledJobToFormInput(initialValues);
        }

        return createEmptyAppScheduledJobFormDefaults();
    }, [initialValues]);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues,
        resolver: zodResolver(CreateOrEditAppScheduledJobFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = methods;
    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, readOnly]);

    const { sources: notificationSources, manageLink: notificationManageLink } =
        useProjectNotificationSettingsSources(projectId);
    const consoleWidthInputId = useId();
    const consoleHeightInputId = useId();

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ control, name: "name" });
    const { field: scheduleMode } = useController({ control, name: "scheduleMode" });
    const {
        field: scheduleInterval,
        fieldState: { invalid: isScheduleIntervalInvalid },
    } = useController({ control, name: "scheduleInterval" });
    const {
        field: scheduleCronExpr,
        fieldState: { invalid: isScheduleCronExprInvalid },
    } = useController({ control, name: "scheduleCronExpr" });
    const {
        field: scheduleFrom,
        fieldState: { invalid: isScheduleFromInvalid },
    } = useController({ control, name: "scheduleFrom" });
    const {
        field: timeout,
        fieldState: { invalid: isTimeoutInvalid },
    } = useController({ control, name: "timeout" });
    const {
        field: maxRetry,
        fieldState: { invalid: isMaxRetryInvalid },
    } = useController({ control, name: "maxRetry" });
    const {
        field: retryDelay,
        fieldState: { invalid: isRetryDelayInvalid },
    } = useController({ control, name: "retryDelay" });
    const { field: priority } = useController({ control, name: "priority" });
    const { field: controlEnabled } = useController({ control, name: "controlEnabled" });
    const { field: runInShell } = useController({ control, name: "runInShell" });
    const { field: commandMode } = useController({ control, name: "commandMode" });
    const {
        field: command,
        fieldState: { invalid: isCommandInvalid },
    } = useController({ control, name: "command" });
    const {
        field: script,
        fieldState: { invalid: isScriptInvalid },
    } = useController({ control, name: "script" });
    const {
        field: workingDir,
        fieldState: { invalid: isWorkingDirInvalid },
    } = useController({ control, name: "workingDir" });
    const { field: tty } = useController({ control, name: "tty" });
    const {
        field: consoleWidth,
        fieldState: { invalid: isConsoleWidthInvalid },
    } = useController({ control, name: "consoleSize.width" });
    const {
        field: consoleHeight,
        fieldState: { invalid: isConsoleHeightInvalid },
    } = useController({ control, name: "consoleSize.height" });

    function onValid(values: SchemaOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<SchemaOutput>) {
        console.error(_errors);
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <fieldset className="contents">
                    <DialogBody>
                        <input
                            type="hidden"
                            {...runInShell}
                            value={runInShell.value}
                        />
                        <FieldGroup className="gap-6">
                            <InfoBlock
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Name"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Input
                                        {...name}
                                        placeholder="scheduled job name"
                                        aria-invalid={isNameInvalid}
                                        className="max-w-[420px]"
                                        disabled={readOnly}
                                    />
                                    <FieldError errors={[errors.name]} />
                                </Field>
                            </InfoBlock>

                            <ContentBlock label="Scheduling">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        title="Priority"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <PriorityTabsField
                                            value={priority.value}
                                            onChange={priority.onChange}
                                            readOnly={readOnly}
                                        />
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Scheduling Mode"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Tabs
                                            value={scheduleMode.value}
                                            onValueChange={scheduleMode.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger
                                                    value={EAppScheduledJobScheduleMode.Interval}
                                                    disabled={readOnly}
                                                >
                                                    Interval-based
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={EAppScheduledJobScheduleMode.Cron}
                                                    disabled={readOnly}
                                                >
                                                    Time-based
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                    </InfoBlock>

                                    {scheduleMode.value === EAppScheduledJobScheduleMode.Interval && (
                                        <InfoBlock
                                            title="Scheduling Interval"
                                            titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                        >
                                            <Field>
                                                <Input
                                                    {...scheduleInterval}
                                                    placeholder="1d, 1h30m"
                                                    className="max-w-[400px]"
                                                    aria-invalid={isScheduleIntervalInvalid}
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.scheduleInterval]} />
                                            </Field>
                                        </InfoBlock>
                                    )}

                                    {scheduleMode.value === EAppScheduledJobScheduleMode.Cron && (
                                        <InfoBlock
                                            title="Cron Expression"
                                            titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                        >
                                            <Field>
                                                <Input
                                                    {...scheduleCronExpr}
                                                    placeholder="accepted form: * * * * *"
                                                    className="max-w-[400px]"
                                                    aria-invalid={isScheduleCronExprInvalid}
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.scheduleCronExpr]} />
                                            </Field>
                                        </InfoBlock>
                                    )}

                                    <InfoBlock
                                        title="Schedule From"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
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
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.scheduleFrom]} />
                                        </Field>
                                    </InfoBlock>

                                    <NextRunsField nextRuns={initialValues?.nextRuns ?? []} />

                                    <InfoBlock
                                        title="Timeout"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Field>
                                            <Input
                                                {...timeout}
                                                placeholder="30m, 1h30m"
                                                className="max-w-[400px]"
                                                aria-invalid={isTimeoutInvalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.timeout]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Max Retry"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Field>
                                            <InputNumber
                                                value={maxRetry.value}
                                                onValueChange={value => {
                                                    const nextValue =
                                                        value !== undefined && Number.isFinite(value)
                                                            ? value
                                                            : undefined;

                                                    maxRetry.onChange(nextValue);
                                                }}
                                                min={0}
                                                useGrouping={false}
                                                placeholder="3"
                                                aria-invalid={isMaxRetryInvalid}
                                                className="max-w-[100px]"
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.maxRetry]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Retry Delay"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Field>
                                            <Input
                                                {...retryDelay}
                                                placeholder="10s"
                                                className="max-w-[400px]"
                                                aria-invalid={isRetryDelayInvalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.retryDelay]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Control Enabled"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Checkbox
                                            checked={controlEnabled.value}
                                            onCheckedChange={checked => {
                                                controlEnabled.onChange(checked === true);
                                            }}
                                            disabled={readOnly}
                                        />
                                    </InfoBlock>
                                </div>
                            </ContentBlock>

                            <ContentBlock label="Command">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        title="Mode"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Tabs
                                            value={commandMode.value}
                                            onValueChange={commandMode.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger
                                                    value={APP_SCHEDULED_JOB_COMMAND_MODE.Command}
                                                    disabled={readOnly}
                                                >
                                                    Command
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={APP_SCHEDULED_JOB_COMMAND_MODE.Script}
                                                    disabled={readOnly}
                                                >
                                                    Script
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                    </InfoBlock>

                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label={
                                                    commandMode.value === APP_SCHEDULED_JOB_COMMAND_MODE.Script
                                                        ? "Script"
                                                        : "Command"
                                                }
                                                isRequired
                                            />
                                        }
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        {commandMode.value === APP_SCHEDULED_JOB_COMMAND_MODE.Script ? (
                                            <Field>
                                                <ScriptEditorField
                                                    value={script.value}
                                                    onChange={script.onChange}
                                                    invalid={isScriptInvalid}
                                                    error={errors.script}
                                                    readOnly={readOnly}
                                                />
                                            </Field>
                                        ) : (
                                            <Field>
                                                <Input
                                                    {...command}
                                                    placeholder="echo “$CMD_ARG_GROUP_1”"
                                                    className="max-w-[400px]"
                                                    aria-invalid={isCommandInvalid}
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.command]} />
                                            </Field>
                                        )}
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Working Directory"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Field>
                                            <Input
                                                {...workingDir}
                                                placeholder="/path/in/container"
                                                className="max-w-[400px]"
                                                aria-invalid={isWorkingDirInvalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.workingDir]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Terminal"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <div className="flex w-full max-w-[400px] flex-wrap items-start gap-x-4 gap-y-3">
                                            <div className="flex h-9 items-center gap-3 text-sm font-medium">
                                                <span>TTY</span>
                                                <Checkbox
                                                    checked={tty.value}
                                                    onCheckedChange={checked => {
                                                        tty.onChange(checked === true);
                                                    }}
                                                    aria-label="TTY"
                                                    disabled={readOnly}
                                                />
                                            </div>

                                            <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <label
                                                        htmlFor={consoleWidthInputId}
                                                        className="shrink-0 text-sm font-medium"
                                                    >
                                                        Width
                                                    </label>
                                                    <InputNumber
                                                        id={consoleWidthInputId}
                                                        ref={consoleWidth.ref}
                                                        name={consoleWidth.name}
                                                        value={consoleWidth.value}
                                                        onBlur={consoleWidth.onBlur}
                                                        onValueChange={value => {
                                                            const nextValue =
                                                                value !== undefined && Number.isFinite(value)
                                                                    ? value
                                                                    : undefined;

                                                            consoleWidth.onChange(nextValue);
                                                        }}
                                                        useGrouping={false}
                                                        showControls={false}
                                                        placeholder="120"
                                                        aria-invalid={isConsoleWidthInvalid}
                                                        className="min-w-0 flex-1"
                                                        disabled={readOnly}
                                                    />
                                                </div>
                                                <FieldError errors={[errors.consoleSize?.width]} />
                                            </div>

                                            <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <label
                                                        htmlFor={consoleHeightInputId}
                                                        className="shrink-0 text-sm font-medium"
                                                    >
                                                        Height
                                                    </label>
                                                    <InputNumber
                                                        id={consoleHeightInputId}
                                                        ref={consoleHeight.ref}
                                                        name={consoleHeight.name}
                                                        value={consoleHeight.value}
                                                        onBlur={consoleHeight.onBlur}
                                                        onValueChange={value => {
                                                            const nextValue =
                                                                value !== undefined && Number.isFinite(value)
                                                                    ? value
                                                                    : undefined;

                                                            consoleHeight.onChange(nextValue);
                                                        }}
                                                        useGrouping={false}
                                                        showControls={false}
                                                        placeholder="40"
                                                        aria-invalid={isConsoleHeightInvalid}
                                                        className="min-w-0 flex-1"
                                                        disabled={readOnly}
                                                    />
                                                </div>
                                                <FieldError errors={[errors.consoleSize?.height]} />
                                            </div>
                                        </div>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Environment Variables"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <KeyValueList<SchemaInput>
                                            name="envVars"
                                            keyLabel="Key"
                                            valueLabel="Value"
                                            className="max-w-[660px]"
                                            checkDuplicates
                                            disabled={readOnly}
                                        />
                                    </InfoBlock>
                                </div>
                            </ContentBlock>

                            <ContentBlock label="Arg Groups">
                                <ArgGroupsSection readOnly={readOnly} />
                            </ContentBlock>

                            <ContentBlock label="Notification Configuration">
                                <NotificationSettings<SchemaInput>
                                    names={{
                                        successUseDefault: "notification.successUseDefault",
                                        success: "notification.success",
                                        failureUseDefault: "notification.failureUseDefault",
                                        failure: "notification.failure",
                                    }}
                                    sources={notificationSources}
                                    manageLink={notificationManageLink}
                                    readOnly={readOnly}
                                    titleWidth={220}
                                />
                            </ContentBlock>
                        </FieldGroup>
                    </DialogBody>
                    <DialogActionFooter>
                        <Button
                            type="submit"
                            isLoading={isPending}
                            disabled={readOnly}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </DialogActionFooter>
                </fieldset>
            </form>
        </FormProvider>
    );
}

interface Props {
    projectId: string;
    isPending: boolean;
    onSubmit: (values: SchemaOutput) => Promise<void> | void;
    initialValues?: AppScheduledJob;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
}
