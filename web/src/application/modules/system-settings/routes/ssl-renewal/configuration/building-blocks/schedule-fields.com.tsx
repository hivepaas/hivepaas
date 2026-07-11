import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import {
    type SystemSslRenewalConfigurationFormInput,
    type SystemSslRenewalConfigurationFormOutput,
    SystemSslRenewalScheduleMode,
} from "../schemas";

type SchemaInput = SystemSslRenewalConfigurationFormInput;
type SchemaOutput = SystemSslRenewalConfigurationFormOutput;

export function ScheduleFields() {
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
            {scheduleMode.value === SystemSslRenewalScheduleMode.Interval && (
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

            {scheduleMode.value === SystemSslRenewalScheduleMode.Cron && (
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
        </>
    );
}
