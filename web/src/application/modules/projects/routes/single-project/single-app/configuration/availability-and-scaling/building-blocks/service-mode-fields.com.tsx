import { FieldError } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { EServiceMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigAvailabilitySchemaInput, type AppConfigAvailabilitySchemaOutput } from "../schemas";

export function ServiceModeFields() {
    const { control } = useFormContext<AppConfigAvailabilitySchemaInput, unknown, AppConfigAvailabilitySchemaOutput>();

    const mode = useWatch({ control, name: "mode" });

    const { field: modeField } = useController({ control, name: "mode" });
    const {
        field: serviceReplicas,
        fieldState: { error: serviceReplicasError },
    } = useController({ control, name: "serviceReplicas" });
    const {
        field: jobMaxConcurrent,
        fieldState: { error: jobMaxConcurrentError },
    } = useController({ control, name: "jobMaxConcurrent" });
    const {
        field: jobTotalCompletions,
        fieldState: { error: jobTotalCompletionsError },
    } = useController({ control, name: "jobTotalCompletions" });

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Service Mode"
                        content="Specifies how the service should be deployed and scaled."
                    />
                }
            >
                <Tabs
                    value={modeField.value}
                    onValueChange={v => {
                        modeField.onChange(v as EServiceMode);
                    }}
                    className="w-fit"
                >
                    <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                        <TabsTrigger value={EServiceMode.Replicated}>Replicated</TabsTrigger>
                        <TabsTrigger value={EServiceMode.ReplicatedJob}>Replicated Job</TabsTrigger>
                        <TabsTrigger value={EServiceMode.Global}>Global</TabsTrigger>
                        <TabsTrigger value={EServiceMode.GlobalJob}>Global Job</TabsTrigger>
                    </TabsList>
                </Tabs>
            </InfoBlock>

            {mode === EServiceMode.Replicated && (
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Replicas"
                            content="Number of instances of the service to run simultaneously."
                        />
                    }
                >
                    <InputNumber
                        value={serviceReplicas.value ?? undefined}
                        onValueChange={val => {
                            serviceReplicas.onChange(val);
                        }}
                        className="max-w-[65px]"
                        min={0}
                    />
                    <FieldError errors={[serviceReplicasError]} />
                </InfoBlock>
            )}

            {mode === EServiceMode.ReplicatedJob && (
                <>
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Max Concurrent"
                                content="Maximum number of concurrently running job instances."
                            />
                        }
                    >
                        <InputNumber
                            value={jobMaxConcurrent.value ?? undefined}
                            onValueChange={val => {
                                jobMaxConcurrent.onChange(val);
                            }}
                            className="max-w-[65px]"
                            min={0}
                        />
                        <FieldError errors={[jobMaxConcurrentError]} />
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Total Completions"
                                content="Total number of successful completions required for the job to be considered complete."
                            />
                        }
                    >
                        <InputNumber
                            value={jobTotalCompletions.value ?? undefined}
                            onValueChange={val => {
                                jobTotalCompletions.onChange(val);
                            }}
                            className="max-w-[65px]"
                            min={0}
                        />
                        <FieldError errors={[jobTotalCompletionsError]} />
                    </InfoBlock>
                </>
            )}
        </div>
    );
}
