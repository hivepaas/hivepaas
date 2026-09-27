import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EAppScheduledJobScheduleMode } from "~/projects/module-shared/enums";

import { InfoBlock, NextRunsField } from "@application/shared/components";

import { Field, FieldError, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import type { JobScheduleFormValues } from "./job-schedule.helpers";

const NO_SCHEDULE_NOTE = "The job runs only when run by hand, or as a step of a job sequence.";

/**
 * A job's schedule, in a form whose values include JobScheduleFormValues: the
 * mode (None, Interval, Cron), the interval or cron expression, the window and
 * the next runs.
 */
export function JobScheduleFields({ titleWidth, nextRuns, readOnly = false }: Props) {
    const { control } = useFormContext<JobScheduleFormValues>();

    const { field: scheduleMode } = useController({ control, name: "scheduleMode" });
    const { field: scheduleInterval, fieldState: intervalState } = useController({ control, name: "scheduleInterval" });
    const { field: scheduleCronExpr, fieldState: cronExprState } = useController({ control, name: "scheduleCronExpr" });
    const { field: scheduleFrom, fieldState: fromState } = useController({ control, name: "scheduleFrom" });
    const { field: scheduleTo, fieldState: toState } = useController({ control, name: "scheduleTo" });

    const mode = scheduleMode.value;

    return (
        <>
            <InfoBlock
                title="Scheduling Mode"
                titleWidth={titleWidth}
            >
                <div className="flex flex-col gap-2">
                    <Tabs
                        value={mode}
                        onValueChange={scheduleMode.onChange}
                    >
                        <TabsList>
                            <TabsTrigger
                                value={EAppScheduledJobScheduleMode.None}
                                disabled={readOnly}
                            >
                                No schedule
                            </TabsTrigger>
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
                    {mode === EAppScheduledJobScheduleMode.None && (
                        <p className="text-sm text-muted-foreground">{NO_SCHEDULE_NOTE}</p>
                    )}
                </div>
            </InfoBlock>

            {mode === EAppScheduledJobScheduleMode.Interval && (
                <InfoBlock
                    title="Scheduling Interval"
                    titleWidth={titleWidth}
                >
                    <Field>
                        <Input
                            {...scheduleInterval}
                            placeholder="1d, 1h30m"
                            className="max-w-[400px]"
                            aria-invalid={intervalState.invalid}
                            disabled={readOnly}
                        />
                        <FieldError errors={[intervalState.error]} />
                    </Field>
                </InfoBlock>
            )}

            {mode === EAppScheduledJobScheduleMode.Cron && (
                <InfoBlock
                    title="Cron Expression"
                    titleWidth={titleWidth}
                >
                    <Field>
                        <Input
                            {...scheduleCronExpr}
                            placeholder="accepted form: * * * * *"
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            aria-invalid={cronExprState.invalid}
                            disabled={readOnly}
                        />
                        <FieldError errors={[cronExprState.error]} />
                    </Field>
                </InfoBlock>
            )}

            {mode !== EAppScheduledJobScheduleMode.None && (
                <>
                    <InfoBlock
                        title="Schedule From"
                        titleWidth={titleWidth}
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
                                    aria-invalid={fromState.invalid}
                                    containerClassName="w-full"
                                    disabled={readOnly}
                                />
                                <FieldError errors={[fromState.error]} />
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
                                    aria-invalid={toState.invalid}
                                    containerClassName="w-full"
                                    disabled={readOnly}
                                />
                                <FieldError errors={[toState.error]} />
                            </Field>
                        </div>
                    </InfoBlock>

                    <NextRunsField
                        nextRuns={nextRuns}
                        titleWidth={titleWidth}
                    />
                </>
            )}
        </>
    );
}

interface Props {
    titleWidth: number;
    nextRuns: Date[];
    readOnly?: boolean;
}
