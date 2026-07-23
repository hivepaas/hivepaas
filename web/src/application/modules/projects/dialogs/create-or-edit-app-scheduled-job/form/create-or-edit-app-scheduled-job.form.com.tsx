import React, { useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useFormState } from "react-hook-form";
import { useToggle, useUpdateEffect } from "react-use";
import type { AppScheduledJob } from "~/projects/domain";
import { CommandArgGroupsSection, CommandConfigSection } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EAppScheduledJobCommandOutputMode, EAppScheduledJobScheduleMode } from "~/projects/module-shared/enums";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { ContentBlock, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { NotificationSettings } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { InputNumber } from "@/components/ui/input-number";

import type { CreateOrEditAppScheduledJobFormInput, CreateOrEditAppScheduledJobFormOutput } from "../schemas";
import { CreateOrEditAppScheduledJobFormSchema } from "../schemas";

import {
    CommandRunInTargetApp,
    INFO_BLOCK_TITLE_WIDTH,
    NextRunsField,
    PipeToAppSection,
    PriorityTabsField,
    SaveToFileSection,
} from "./building-blocks";
import {
    createEmptyAppScheduledJobFormDefaults,
    mapAppScheduledJobToFormInput,
} from "./create-or-edit-app-scheduled-job.form-mappers";

type SchemaInput = CreateOrEditAppScheduledJobFormInput;
type SchemaOutput = CreateOrEditAppScheduledJobFormOutput;

export function CreateOrEditAppScheduledJobForm({
    projectId,
    appId,
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
            return mapAppScheduledJobToFormInput(initialValues, projectId);
        }

        return createEmptyAppScheduledJobFormDefaults(projectId);
    }, [initialValues, projectId]);

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
    const [isRetryConfigVisible, toggleRetryConfigVisible] = useToggle(false);
    const retryMaxInputId = React.useId();
    const retryDelayInputId = React.useId();
    const retryDelayIncrInputId = React.useId();
    const retryBackoffInputId = React.useId();
    const retryDelayMaxInputId = React.useId();

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
    const { field: commandOutputMode } = useController({ control, name: "commandOutputMode" });

    const configureTemplatesLink = ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId);

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
                                        {!isRetryConfigVisible ? (
                                            <Button
                                                type="button"
                                                variant="link"
                                                className="h-auto p-0 text-sm text-primary w-fit"
                                                onClick={toggleRetryConfigVisible}
                                            >
                                                Show Configuration
                                            </Button>
                                        ) : (
                                            <div className="flex w-full max-w-[1180px] gap-3">
                                                <Button
                                                    type="button"
                                                    variant="link"
                                                    className="h-auto w-fit p-0 text-sm text-primary"
                                                    onClick={toggleRetryConfigVisible}
                                                >
                                                    Hide Configuration
                                                </Button>
                                                <div className="flex w-full flex-wrap items-start gap-x-5 gap-y-3">
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
                                            </div>
                                        )}
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

                            <CommandConfigSection
                                label="Command"
                                showLoadTemplate
                                templateProjectId={projectId}
                                configureTemplatesLink={configureTemplatesLink}
                                readOnly={readOnly}
                            />

                            <ContentBlock label="Arg Groups">
                                <CommandArgGroupsSection
                                    readOnly={readOnly}
                                    fieldName="argGroups"
                                />
                            </ContentBlock>

                            <ContentBlock label="Command Output">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        title="Output"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Tabs
                                            value={commandOutputMode.value}
                                            onValueChange={commandOutputMode.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger
                                                    value={EAppScheduledJobCommandOutputMode.Ignore}
                                                    disabled={readOnly}
                                                >
                                                    Ignore
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={EAppScheduledJobCommandOutputMode.SaveToFile}
                                                    disabled={readOnly}
                                                >
                                                    Save to File
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={EAppScheduledJobCommandOutputMode.PipeToApp}
                                                    disabled={readOnly}
                                                >
                                                    Pipe to App
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                    </InfoBlock>
                                </div>
                            </ContentBlock>

                            {commandOutputMode.value === EAppScheduledJobCommandOutputMode.SaveToFile && (
                                <SaveToFileSection
                                    projectId={projectId}
                                    readOnly={readOnly}
                                />
                            )}

                            {commandOutputMode.value === EAppScheduledJobCommandOutputMode.PipeToApp && (
                                <>
                                    <PipeToAppSection
                                        projectId={projectId}
                                        appId={appId}
                                        readOnly={readOnly}
                                    />
                                    <CommandRunInTargetApp
                                        projectId={projectId}
                                        appId={appId}
                                        readOnly={readOnly}
                                    />
                                </>
                            )}

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
                                className="min-w-[100px]"
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
    appId: string;
    isPending: boolean;
    onSubmit: (values: SchemaOutput) => Promise<void> | void;
    initialValues?: AppScheduledJob;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}
