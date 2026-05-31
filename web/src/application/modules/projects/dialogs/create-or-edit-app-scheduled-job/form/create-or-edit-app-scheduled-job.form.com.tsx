import React, { useMemo } from "react";

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
import { InputNumber } from "@/components/ui/input-number";

import type { CreateOrEditAppScheduledJobFormInput, CreateOrEditAppScheduledJobFormOutput } from "../schemas";
import { CreateOrEditAppScheduledJobFormSchema } from "../schemas";

import {
    ArgGroupsSection,
    INFO_BLOCK_TITLE_WIDTH,
    NextRunsField,
    PriorityTabsField,
    SectionHeader,
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
    const {
        field: command,
        fieldState: { invalid: isCommandInvalid },
    } = useController({ control, name: "command" });
    const {
        field: workingDir,
        fieldState: { invalid: isWorkingDirInvalid },
    } = useController({ control, name: "workingDir" });

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
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
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
                                            <TabsTrigger value={EAppScheduledJobScheduleMode.Interval}>
                                                Interval-based
                                            </TabsTrigger>
                                            <TabsTrigger value={EAppScheduledJobScheduleMode.Cron}>
                                                Time-based
                                            </TabsTrigger>
                                        </TabsList>
                                    </Tabs>
                                </InfoBlock>

                                {scheduleMode.value === EAppScheduledJobScheduleMode.Interval && (
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="Scheduling Interval"
                                                isRequired
                                            />
                                        }
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
                                        title={
                                            <LabelWithInfo
                                                label="Cron Expression"
                                                isRequired
                                            />
                                        }
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
                                    title={
                                        <LabelWithInfo
                                            label="Timeout"
                                            isRequired
                                        />
                                    }
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
                                    title={
                                        <LabelWithInfo
                                            label="Max Retry"
                                            isRequired
                                        />
                                    }
                                    titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                >
                                    <Field>
                                        <InputNumber
                                            value={maxRetry.value}
                                            onValueChange={value => {
                                                maxRetry.onChange(value ?? 0);
                                            }}
                                            min={0}
                                            useGrouping={false}
                                            aria-invalid={isMaxRetryInvalid}
                                            className="max-w-[260px]"
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.maxRetry]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Retry Delay"
                                            isRequired
                                        />
                                    }
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
                                    title={
                                        <LabelWithInfo
                                            label="Command"
                                            isRequired
                                        />
                                    }
                                    titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                >
                                    <Field>
                                        <Input
                                            {...command}
                                            placeholder='my_app --arg1=123 --arg2="my data"'
                                            className="max-w-[520px]"
                                            aria-invalid={isCommandInvalid}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.command]} />
                                    </Field>
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

                                <SectionHeader>Arg Groups</SectionHeader>
                                <ArgGroupsSection readOnly={readOnly} />
                            </div>
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
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            />
                        </ContentBlock>

                        <div className="flex justify-end">
                            <Button
                                type="submit"
                                isLoading={isPending}
                                disabled={readOnly}
                                className="min-w-[140px]"
                            >
                                Save
                            </Button>
                        </div>
                    </FieldGroup>
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
