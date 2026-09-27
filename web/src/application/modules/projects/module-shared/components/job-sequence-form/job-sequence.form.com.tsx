import { useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { FormProvider, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppScheduledJob } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { ESchedJobSeqMode, ESchedJobSeqOnFailure } from "~/projects/module-shared/enums";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { ContentBlock, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { NotificationSettings } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import { JobScheduleFields, PriorityTabsField } from "../job-schedule-fields";
import { JobTriggersField } from "../job-triggers-field";

import { type JobSequenceCandidate, JobSequenceStepsField } from "./job-sequence-steps-field.com";
import { createEmptyJobSequenceFormDefaults, mapJobSequenceToFormInput } from "./job-sequence.form-mappers";
import {
    type JobSequenceFormInput,
    type JobSequenceFormOutput,
    JobSequenceFormSchema,
} from "./job-sequence.form.schema";

const TITLE_WIDTH = 220;

/** What a step's command is told, and how it hands values to the steps after it. */
function StepEnvironmentHelp() {
    return (
        <div className={cn(dashedBorderBox, "flex max-w-[720px] flex-col gap-2 text-sm leading-normal")}>
            <p>
                A step whose job runs a command is told how the steps before it went, in its environment:{" "}
                <code>HIVEPAAS_SEQ_STEP</code> and <code>HIVEPAAS_SEQ_STEPS</code> (its number and the count),{" "}
                <code>HIVEPAAS_SEQ_PREV_STATUS</code> (<code>done</code>, <code>failed</code> or <code>skipped</code>),{" "}
                <code>HIVEPAAS_SEQ_RESULTS</code> (every earlier step, as JSON) and <code>HIVEPAAS_SEQ_OUTPUTS</code>.
            </p>
            <p>
                To hand a value on, a step writes <code>KEY=value</code> lines to the file named by{" "}
                <code>$HIVEPAAS_OUTPUT</code>, for example{" "}
                <code>echo &quot;VERSION=1.2.3&quot; &gt;&gt; &quot;$HIVEPAAS_OUTPUT&quot;</code>. The steps after it
                read it as <code>HIVEPAAS_SEQ_OUTPUT_VERSION</code>. Up to 64 KB, one line per value.
            </p>
            <p>
                <span className="font-semibold text-orange-500">Note:</span> each step retries and times out as its job
                does. A step the server was running when it restarted runs again: make the jobs safe to run twice.
            </p>
        </div>
    );
}

/**
 * A job sequence, at the app or the env scope: its steps are picked from the
 * candidates the caller loads for its scope.
 */
export function JobSequenceForm({
    projectId,
    env,
    candidates,
    isLoadingCandidates,
    triggerApps,
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    readOnly = false,
    stickyActions = false,
    onClose,
}: Props) {
    const defaultValues = useMemo(
        () => (initialValues ? mapJobSequenceToFormInput(initialValues) : createEmptyJobSequenceFormDefaults()),
        [initialValues],
    );

    const methods = useForm<JobSequenceFormInput, unknown, JobSequenceFormOutput>({
        defaultValues,
        resolver: zodResolver(JobSequenceFormSchema),
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
    const { field: mode } = useController({ control, name: "mode" });
    const { field: onFailure } = useController({ control, name: "onFailure" });
    const { field: timeout, fieldState: timeoutState } = useController({ control, name: "timeout" });
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
                                    placeholder="job sequence name"
                                    aria-invalid={nameState.invalid}
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </InfoBlock>

                        <ContentBlock label="Steps">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Jobs"
                                            isRequired
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <JobSequenceStepsField
                                        candidates={candidates}
                                        isLoadingCandidates={isLoadingCandidates}
                                        readOnly={readOnly}
                                    />
                                </InfoBlock>

                                <InfoBlock
                                    title="Mode"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Tabs
                                        value={mode.value}
                                        onValueChange={mode.onChange}
                                    >
                                        <TabsList>
                                            <TabsTrigger
                                                value={ESchedJobSeqMode.Sequential}
                                                disabled={readOnly}
                                            >
                                                Sequential
                                            </TabsTrigger>
                                            <TabsTrigger
                                                value="parallel"
                                                disabled
                                                title="Coming later"
                                            >
                                                Parallel (coming)
                                            </TabsTrigger>
                                        </TabsList>
                                    </Tabs>
                                </InfoBlock>

                                <InfoBlock
                                    title="On Failure"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <div className="flex flex-col gap-2">
                                        <Tabs
                                            value={onFailure.value}
                                            onValueChange={onFailure.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger
                                                    value={ESchedJobSeqOnFailure.Stop}
                                                    disabled={readOnly}
                                                >
                                                    Stop
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value={ESchedJobSeqOnFailure.Continue}
                                                    disabled={readOnly}
                                                >
                                                    Continue
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                        <p className="text-sm text-muted-foreground">
                                            {onFailure.value === ESchedJobSeqOnFailure.Stop
                                                ? "A step that fails ends the run; the steps after it are skipped."
                                                : "Every step runs; the run fails if one of them did."}
                                        </p>
                                    </div>
                                </InfoBlock>

                                <StepEnvironmentHelp />
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
                                            label="Step Timeout"
                                            content="For a step whose job has no timeout of its own."
                                        />
                                    }
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Field>
                                        <Input
                                            {...timeout}
                                            placeholder="30m, 1h30m"
                                            className="max-w-[400px]"
                                            aria-invalid={timeoutState.invalid}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.timeout]} />
                                    </Field>
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

                        <ContentBlock label="Triggers">
                            <JobTriggersField
                                apps={triggerApps}
                                readOnly={readOnly}
                            />
                        </ContentBlock>

                        <ContentBlock label="Notification Configuration">
                            <NotificationSettings<JobSequenceFormInput>
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
    env?: string;
    candidates: JobSequenceCandidate[];
    isLoadingCandidates: boolean;
    /** The env's apps a trigger may name; undefined for an app's sequence, which listens to its app. */
    triggerApps?: { id: string; name: string }[];
    isPending: boolean;
    onSubmit: (values: JobSequenceFormOutput) => void;
    initialValues?: AppScheduledJob;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}
