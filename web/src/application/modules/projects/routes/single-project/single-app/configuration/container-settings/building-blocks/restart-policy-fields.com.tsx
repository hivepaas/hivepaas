import { FieldError, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { useController, useFormContext } from "react-hook-form";
import { ERestartPolicyCondition } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

export function RestartPolicyFields() {
    const { control } = useFormContext<
        AppConfigContainerSettingsFormSchemaInput,
        unknown,
        AppConfigContainerSettingsFormSchemaOutput
    >();

    const { field: condition } = useController({ control, name: "restartPolicy.condition" });
    const {
        field: delay,
        fieldState: { error: delayError },
    } = useController({
        control,
        name: "restartPolicy.delay",
    });
    const {
        field: window,
        fieldState: { error: windowError },
    } = useController({
        control,
        name: "restartPolicy.window",
    });
    const {
        field: maxAttempts,
        fieldState: { error: maxAttemptsError },
    } = useController({ control, name: "restartPolicy.maxAttempts" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Condition"
                        content="When Docker should restart the container if it exits."
                    />
                }
            >
                <Tabs
                    value={condition.value}
                    onValueChange={v => {
                        condition.onChange(v as ERestartPolicyCondition);
                    }}
                    className="w-fit"
                >
                    <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                        <TabsTrigger value={ERestartPolicyCondition.None}>None</TabsTrigger>
                        <TabsTrigger value={ERestartPolicyCondition.OnFailure}>On Failure</TabsTrigger>
                        <TabsTrigger value={ERestartPolicyCondition.Any}>Any</TabsTrigger>
                    </TabsList>
                </Tabs>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Restart Delay"
                        content="Delay before restarting after the container exits."
                    />
                }
            >
                <Input
                    {...delay}
                    value={delay.value}
                    onChange={delay.onChange}
                    placeholder="5s"
                    className="max-w-[400px]"
                />
                <FieldError errors={[delayError]} />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Restart Max Attempts"
                        content="Maximum number of restart attempts."
                    />
                }
            >
                <InputNumber
                    value={maxAttempts.value}
                    onValueChange={value => {
                        maxAttempts.onChange(value);
                    }}
                    placeholder="3"
                    className="max-w-[400px]"
                    min={0}
                    useGrouping={false}
                />
                <FieldError errors={[maxAttemptsError]} />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Restart Window"
                        content="Time window used together with max attempts."
                    />
                }
            >
                <Input
                    {...window}
                    value={window.value}
                    onChange={window.onChange}
                    placeholder="5s"
                    className="max-w-[400px]"
                />
                <FieldError errors={[windowError]} />
            </InfoBlock>
        </div>
    );
}
