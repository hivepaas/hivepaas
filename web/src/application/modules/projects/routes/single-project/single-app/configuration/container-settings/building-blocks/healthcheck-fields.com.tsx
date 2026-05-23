import { Checkbox, FieldError, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { EHealthcheckMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

export function HealthcheckFields() {
    const { control } = useFormContext<
        AppConfigContainerSettingsFormSchemaInput,
        unknown,
        AppConfigContainerSettingsFormSchemaOutput
    >();

    const { field: enabled } = useController({ control, name: "healthcheck.enabled" });
    const { field: mode } = useController({ control, name: "healthcheck.mode" });
    const {
        field: command,
        fieldState: { error: commandError },
    } = useController({ control, name: "healthcheck.command" });
    const {
        field: interval,
        fieldState: { error: intervalError },
    } = useController({ control, name: "healthcheck.interval" });
    const {
        field: timeout,
        fieldState: { error: timeoutError },
    } = useController({ control, name: "healthcheck.timeout" });
    const {
        field: startPeriod,
        fieldState: { error: startPeriodError },
    } = useController({ control, name: "healthcheck.startPeriod" });
    const {
        field: startInterval,
        fieldState: { error: startIntervalError },
    } = useController({ control, name: "healthcheck.startInterval" });
    const {
        field: retries,
        fieldState: { error: retriesError },
    } = useController({ control, name: "healthcheck.retries" });

    const isEnabled = useWatch({ control, name: "healthcheck.enabled" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock title="Enabled">
                <Checkbox
                    checked={enabled.value}
                    onCheckedChange={v => {
                        enabled.onChange(v === true);
                    }}
                />
            </InfoBlock>

            {isEnabled && (
                <>
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Mode"
                                content="How Docker should execute the healthcheck command."
                            />
                        }
                    >
                        <Tabs
                            value={mode.value}
                            onValueChange={v => {
                                mode.onChange(v as EHealthcheckMode);
                            }}
                            className="w-fit"
                        >
                            <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                                <TabsTrigger value={EHealthcheckMode.Inherit}>Inherit</TabsTrigger>
                                <TabsTrigger value={EHealthcheckMode.Cmd}>CMD</TabsTrigger>
                                <TabsTrigger value={EHealthcheckMode.CmdShell}>CMD-SHELL</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Command"
                                content="Command used by Docker to determine if the container is healthy."
                            />
                        }
                    >
                        <Input
                            {...command}
                            value={command.value}
                            onChange={command.onChange}
                            placeholder="wget 127.0.0.1"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[commandError]} />
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Interval"
                                content="Time between healthcheck executions. Zero or blank inherits Docker defaults."
                            />
                        }
                    >
                        <Input
                            {...interval}
                            value={interval.value}
                            onChange={interval.onChange}
                            placeholder="30s"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[intervalError]} />
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Timeout"
                                content="Maximum time allowed for one healthcheck execution."
                            />
                        }
                    >
                        <Input
                            {...timeout}
                            value={timeout.value}
                            onChange={timeout.onChange}
                            placeholder="5s"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[timeoutError]} />
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Start Period"
                                content="Grace period before healthcheck failures count."
                            />
                        }
                    >
                        <Input
                            {...startPeriod}
                            value={startPeriod.value}
                            onChange={startPeriod.onChange}
                            placeholder="3s"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[startPeriodError]} />
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Start Interval"
                                content="Time between healthchecks during the start period."
                            />
                        }
                    >
                        <Input
                            {...startInterval}
                            value={startInterval.value}
                            onChange={startInterval.onChange}
                            placeholder="3s"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[startIntervalError]} />
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Retries"
                                content="Consecutive failures needed before Docker marks the container unhealthy."
                            />
                        }
                    >
                        <InputNumber
                            value={retries.value}
                            onValueChange={value => {
                                retries.onChange(value);
                            }}
                            placeholder="3"
                            className="max-w-[400px]"
                            min={0}
                            useGrouping={false}
                        />
                        <FieldError errors={[retriesError]} />
                    </InfoBlock>
                </>
            )}
        </div>
    );
}
