import React, { useId, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppScheduledJob } from "~/projects/domain";
import { CommandArgGroupsSection } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EAppScheduledJobScheduleMode } from "~/projects/module-shared/enums";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { ContentBlock, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList, NotificationSettings } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { InputNumber } from "@/components/ui/input-number";

import type { CreateOrEditAppScheduledJobFormInput, CreateOrEditAppScheduledJobFormOutput } from "../schemas";
import { APP_SCHEDULED_JOB_COMMAND_MODE, CreateOrEditAppScheduledJobFormSchema } from "../schemas";

import { INFO_BLOCK_TITLE_WIDTH, NextRunsField, PriorityTabsField, ScriptEditorField } from "./building-blocks";
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
    stickyActions = false,
    onClose,
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
    const retryMaxInputId = useId();
    const retryDelayInputId = useId();
    const retryDelayIncrInputId = useId();
    const retryBackoffInputId = useId();
    const retryDelayMaxInputId = useId();
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
        field: scheduleTo,
        fieldState: { invalid: isScheduleToInvalid },
    } = useController({ control, name: "scheduleTo" });
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
    const {
        field: retryDelayIncr,
        fieldState: { invalid: isRetryDelayIncrInvalid },
    } = useController({ control, name: "retryDelayIncr" });
    const { field: retryBackoff } = useController({ control, name: "retryBackoff" });
    const {
        field: retryDelayMax,
        fieldState: { invalid: isRetryDelayMaxInvalid },
    } = useController({ control, name: "retryDelayMax" });
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
                    <div>
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
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                                        <div className="flex w-full max-w-[600px] flex-wrap items-start gap-x-4 gap-y-3">
                                            <Field className="min-w-[260px] flex-1">
                                                <DateTimePicker
                                                    value={scheduleFrom.value ?? undefined}
                                                    onChange={date => {
                                                        scheduleFrom.onChange(date ?? null);
                                                    }}
                                                    placeholder="select date time"
                                                    granularity="minute"
                                                    showClearButton
                                                    aria-invalid={isScheduleFromInvalid}
                                                    containerClassName="w-full"
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.scheduleFrom]} />
                                            </Field>

                                            <div className="flex h-9 items-center text-sm font-medium">To</div>

                                            <Field className="min-w-[260px] flex-1">
                                                <DateTimePicker
                                                    value={scheduleTo.value ?? undefined}
                                                    onChange={date => {
                                                        scheduleTo.onChange(date ?? null);
                                                    }}
                                                    placeholder="select date time"
                                                    granularity="minute"
                                                    showClearButton
                                                    aria-invalid={isScheduleToInvalid}
                                                    containerClassName="w-full"
                                                    disabled={readOnly}
                                                />
                                                <FieldError errors={[errors.scheduleTo]} />
                                            </Field>
                                        </div>
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
                                        title="Retry"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <div className="flex w-full max-w-[1180px] flex-wrap items-start gap-x-5 gap-y-3">
                                            <div className="flex min-w-[130px] flex-col gap-1.5">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <label
                                                        htmlFor={retryMaxInputId}
                                                        className="shrink-0 text-sm font-medium"
                                                    >
                                                        Max
                                                    </label>
                                                    <InputNumber
                                                        id={retryMaxInputId}
                                                        ref={maxRetry.ref}
                                                        name={maxRetry.name}
                                                        value={maxRetry.value}
                                                        onBlur={maxRetry.onBlur}
                                                        onValueChange={value => {
                                                            const nextValue =
                                                                value !== undefined && Number.isFinite(value)
                                                                    ? value
                                                                    : undefined;

                                                            maxRetry.onChange(nextValue);
                                                        }}
                                                        min={0}
                                                        useGrouping={false}
                                                        showControls={false}
                                                        placeholder="0"
                                                        aria-invalid={isMaxRetryInvalid}
                                                        className="w-[92px]"
                                                        disabled={readOnly}
                                                    />
                                                </div>
                                                <FieldError errors={[errors.maxRetry]} />
                                            </div>

                                            <div className="flex min-w-[0] flex-col gap-1.5">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <label
                                                        htmlFor={retryDelayInputId}
                                                        className="shrink-0 text-sm font-medium"
                                                    >
                                                        Delay
                                                    </label>
                                                    <div className="">
                                                        <Input
                                                            id={retryDelayInputId}
                                                            {...retryDelay}
                                                            placeholder="10s"
                                                            className="w-[110px]"
                                                            aria-invalid={isRetryDelayInvalid}
                                                            disabled={readOnly}
                                                        />
                                                        <FieldError errors={[errors.retryDelay]} />
                                                    </div>
                                                </div>
                                            </div>

                                            <div className="flex min-w-0 flex-col gap-1.5">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <label
                                                        htmlFor={retryDelayIncrInputId}
                                                        className="shrink-0 text-sm font-medium"
                                                    >
                                                        Delay Incr
                                                    </label>
                                                    <Input
                                                        id={retryDelayIncrInputId}
                                                        {...retryDelayIncr}
                                                        placeholder="5s"
                                                        className="w-[110px]"
                                                        aria-invalid={isRetryDelayIncrInvalid}
                                                        disabled={readOnly}
                                                    />
                                                </div>
                                                <FieldError errors={[errors.retryDelayIncr]} />
                                            </div>

                                            <div className="flex h-9 items-center gap-3 text-sm font-medium">
                                                <label htmlFor={retryBackoffInputId}>Expo Backoff</label>
                                                <Checkbox
                                                    id={retryBackoffInputId}
                                                    checked={retryBackoff.value}
                                                    onCheckedChange={checked => {
                                                        retryBackoff.onChange(checked === true);
                                                    }}
                                                    aria-label="Expo Backoff"
                                                    disabled={readOnly}
                                                />
                                            </div>

                                            <div className="flex min-w-0 flex-col gap-1.5">
                                                <div className="flex min-w-0 items-center gap-2">
                                                    <label
                                                        htmlFor={retryDelayMaxInputId}
                                                        className="shrink-0 text-sm font-medium"
                                                    >
                                                        Delay Max
                                                    </label>
                                                    <Input
                                                        id={retryDelayMaxInputId}
                                                        {...retryDelayMax}
                                                        placeholder="24h"
                                                        className="w-[110px]"
                                                        aria-invalid={isRetryDelayMaxInvalid}
                                                        disabled={readOnly}
                                                    />
                                                </div>
                                                <FieldError errors={[errors.retryDelayMax]} />
                                            </div>
                                        </div>
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
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                                            className="max-w-[600px]"
                                            checkDuplicates
                                            disabled={readOnly}
                                        />
                                    </InfoBlock>
                                </div>
                            </ContentBlock>

                            <ContentBlock label="Arg Groups">
                                <CommandArgGroupsSection readOnly={readOnly} />
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
                    </div>
                    {!readOnly && (
                        <FormActionBar sticky={stickyActions}>
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isPending}
                                onClick={onClose}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                isLoading={isPending}
                                className="min-w-[100px]"
                            >
                                Save
                            </Button>
                        </FormActionBar>
                    )}
                    {readOnly && (
                        <FormActionBar sticky={stickyActions}>
                            <Button
                                type="button"
                                onClick={onClose}
                            >
                                Close
                            </Button>
                        </FormActionBar>
                    )}
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
    stickyActions?: boolean;
    onClose?: () => void;
}
