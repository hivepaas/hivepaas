import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import {
    Field,
    FieldError,
    FieldGroup,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import {
    REGISTRY_AUTH_RENEWAL_INTERVALS,
    type SystemRegistryAuthRenewalConfigurationFormInput,
    type SystemRegistryAuthRenewalConfigurationFormOutput,
} from "../schemas";

type SchemaInput = SystemRegistryAuthRenewalConfigurationFormInput;
type SchemaOutput = SystemRegistryAuthRenewalConfigurationFormOutput;

export function ScheduleFields() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field: scheduleInterval,
        fieldState: { error: scheduleIntervalError, invalid: isScheduleIntervalInvalid },
    } = useController({ control, name: "scheduleInterval" });
    const {
        field: scheduleFrom,
        fieldState: { error: scheduleFromError, invalid: isScheduleFromInvalid },
    } = useController({ control, name: "scheduleFrom" });

    // An interval set through the API that is not a whole hour is still shown.
    const intervals: readonly string[] = REGISTRY_AUTH_RENEWAL_INTERVALS.some(value => value === scheduleInterval.value)
        ? REGISTRY_AUTH_RENEWAL_INTERVALS
        : [scheduleInterval.value, ...REGISTRY_AUTH_RENEWAL_INTERVALS];

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title="Interval"
            >
                <FieldGroup>
                    <Field>
                        <Select
                            value={scheduleInterval.value}
                            onValueChange={scheduleInterval.onChange}
                        >
                            <SelectTrigger
                                className="w-full max-w-[400px]"
                                aria-invalid={isScheduleIntervalInvalid}
                            >
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {intervals.map(value => (
                                    <SelectItem
                                        key={value}
                                        value={value}
                                    >
                                        Every {value}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <FieldError errors={[scheduleIntervalError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

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
        </>
    );
}
