import { useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppScheduledJob } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { ContentBlock, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { NotificationSettings } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

import { JobScheduleFields, PriorityTabsField } from "../job-schedule-fields";

import { createEmptyFunctionInvokeFormDefaults, mapFunctionInvokeToFormInput } from "./function-invoke.form-mappers";
import {
    FUNCTION_INVOKE_METHODS,
    type FunctionInvokeFormInput,
    type FunctionInvokeFormOutput,
    FunctionInvokeFormSchema,
    methodSendsNoBody,
} from "./function-invoke.form.schema";

const TITLE_WIDTH = 220;

/**
 * A function's call on a schedule: the request it sends. The runtime's invoke
 * answers it in a running instance of the function; a status of 400 or more
 * fails the run.
 */
export function FunctionInvokeForm({
    projectId,
    env,
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    readOnly = false,
    stickyActions = false,
    onClose,
}: Props) {
    const defaultValues = useMemo(
        () => (initialValues ? mapFunctionInvokeToFormInput(initialValues) : createEmptyFunctionInvokeFormDefaults()),
        [initialValues],
    );

    const methods = useForm<FunctionInvokeFormInput, unknown, FunctionInvokeFormOutput>({
        defaultValues,
        resolver: zodResolver(FunctionInvokeFormSchema),
        mode: "onSubmit",
    });
    const {
        control,
        handleSubmit,
        formState: { errors },
    } = methods;
    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        methods.reset(defaultValues);
    }, [defaultValues]);

    useUpdateEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, readOnly]);

    const { sources: notificationSources, manageLink: notificationManageLink } = useProjectNotificationSettingsSources(
        projectId,
        env,
    );

    const { field: name, fieldState: nameState } = useController({ control, name: "name" });
    const { field: method } = useController({ control, name: "method" });
    const { field: path, fieldState: pathState } = useController({ control, name: "path" });
    const { field: headers, fieldState: headersState } = useController({ control, name: "headers" });
    const { field: body, fieldState: bodyState } = useController({ control, name: "body" });
    const { field: timeout, fieldState: timeoutState } = useController({ control, name: "timeout" });
    const { field: maxRetry, fieldState: maxRetryState } = useController({ control, name: "maxRetry" });
    const { field: retryDelay, fieldState: retryDelayState } = useController({ control, name: "retryDelay" });
    const { field: priority } = useController({ control, name: "priority" });
    const { field: controlEnabled } = useController({ control, name: "controlEnabled" });

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(values => {
                        onSubmit(values);
                    })(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <fieldset className="contents">
                    <FieldGroup className="gap-6">
                        <InfoBlock
                            titleWidth={TITLE_WIDTH}
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
                                    placeholder="nightly report"
                                    aria-invalid={nameState.invalid}
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </InfoBlock>

                        <ContentBlock label="Request">
                            <div className="flex flex-col gap-6">
                                <p className="text-sm text-muted-foreground">
                                    The function answers it in one of its running instances. A status of 400 or more
                                    fails the run; the response is the run&apos;s output.
                                </p>

                                <InfoBlock
                                    title="Method"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Select
                                        value={method.value}
                                        onValueChange={method.onChange}
                                        disabled={readOnly}
                                    >
                                        <SelectTrigger className="w-40">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {FUNCTION_INVOKE_METHODS.map(item => (
                                                <SelectItem
                                                    key={item}
                                                    value={item}
                                                >
                                                    {item}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </InfoBlock>

                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Path"
                                            content="The path the handler sees, with its query."
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Field>
                                        <Input
                                            {...path}
                                            placeholder="/report?day=today"
                                            className={`${PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS} font-mono`}
                                            aria-invalid={pathState.invalid}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.path]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Headers"
                                            content="One per line, as name: value."
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Field>
                                        <Textarea
                                            {...headers}
                                            placeholder="content-type: application/json"
                                            rows={3}
                                            className={`${PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS} font-mono text-xs`}
                                            aria-invalid={headersState.invalid}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.headers]} />
                                    </Field>
                                </InfoBlock>

                                {!methodSendsNoBody(method.value) && (
                                    <InfoBlock
                                        title="Body"
                                        titleWidth={TITLE_WIDTH}
                                    >
                                        <Field>
                                            <Textarea
                                                {...body}
                                                placeholder='{"day":"today"}'
                                                rows={6}
                                                className={`${PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS} font-mono text-xs`}
                                                aria-invalid={bodyState.invalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.body]} />
                                        </Field>
                                    </InfoBlock>
                                )}
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Scheduling">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    title="Priority"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <PriorityTabsField
                                        value={priority.value}
                                        onChange={priority.onChange}
                                        readOnly={readOnly}
                                    />
                                </InfoBlock>

                                <JobScheduleFields
                                    titleWidth={TITLE_WIDTH}
                                    nextRuns={initialValues?.nextRuns ?? []}
                                    readOnly={readOnly}
                                />

                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Timeout"
                                            content="The run's, beside the function's own timeout of a call."
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Field>
                                        <Input
                                            {...timeout}
                                            placeholder="5m, 1h30m"
                                            className="max-w-[400px]"
                                            aria-invalid={timeoutState.invalid}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.timeout]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    title="Retry"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <div className="flex flex-wrap items-start gap-4">
                                        <Field className="w-auto">
                                            <Input
                                                ref={maxRetry.ref}
                                                name={maxRetry.name}
                                                type="number"
                                                inputMode="numeric"
                                                min={0}
                                                value={maxRetry.value ?? ""}
                                                onBlur={maxRetry.onBlur}
                                                onChange={event => {
                                                    const next = Number(event.target.value);
                                                    maxRetry.onChange(
                                                        event.target.value === "" || !Number.isFinite(next)
                                                            ? undefined
                                                            : next,
                                                    );
                                                }}
                                                placeholder="max, 0"
                                                className="w-24"
                                                aria-invalid={maxRetryState.invalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.maxRetry]} />
                                        </Field>
                                        <Field className="w-auto">
                                            <Input
                                                {...retryDelay}
                                                placeholder="delay, 1m"
                                                className="w-24"
                                                aria-invalid={retryDelayState.invalid}
                                                disabled={readOnly}
                                            />
                                            <FieldError errors={[errors.retryDelay]} />
                                        </Field>
                                    </div>
                                </InfoBlock>

                                <InfoBlock
                                    title="Allow Canceling"
                                    titleWidth={TITLE_WIDTH}
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

                        <ContentBlock label="Notification Configuration">
                            <NotificationSettings<FunctionInvokeFormInput>
                                names={{
                                    successUseDefault: "notification.successUseDefault",
                                    success: "notification.success",
                                    failureUseDefault: "notification.failureUseDefault",
                                    failure: "notification.failure",
                                }}
                                sources={notificationSources}
                                manageLink={notificationManageLink}
                                readOnly={readOnly}
                                titleWidth={TITLE_WIDTH}
                            />
                        </ContentBlock>
                    </FieldGroup>

                    <FormActionBar sticky={stickyActions}>
                        {readOnly ? (
                            <Button
                                type="button"
                                onClick={onClose}
                                className="min-w-[100px]"
                            >
                                Close
                            </Button>
                        ) : (
                            <>
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
                            </>
                        )}
                    </FormActionBar>
                </fieldset>
            </form>
        </FormProvider>
    );
}

interface Props {
    projectId: string;
    env: string;
    isPending: boolean;
    onSubmit: (values: FunctionInvokeFormOutput) => void;
    initialValues?: AppScheduledJob;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}
