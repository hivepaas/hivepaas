import { FieldError } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { EServiceMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigAvailabilitySchemaInput, type AppConfigAvailabilitySchemaOutput } from "../schemas";

export function ServiceModeFields({ savedMode, isAppStopped = false }: Props) {
    const { control } = useFormContext<AppConfigAvailabilitySchemaInput, unknown, AppConfigAvailabilitySchemaOutput>();

    const mode = useWatch({ control, name: "mode" });

    // Swarm cannot change the mode of an existing service, so HivePaaS deletes and recreates it.
    const isChangingMode = Boolean(savedMode) && mode !== savedMode;

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
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Service Mode"
                        content="Specifies how the service should be deployed and scaled."
                    />
                }
            >
                <div className="flex flex-col gap-2">
                    <Tabs
                        value={modeField.value}
                        onValueChange={v => {
                            modeField.onChange(v as EServiceMode);
                        }}
                        className="w-fit"
                    >
                        <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                            <TabsTrigger
                                value={EServiceMode.Replicated}
                                disabled={isAppStopped && savedMode !== EServiceMode.Replicated}
                            >
                                Replicated
                            </TabsTrigger>
                            <TabsTrigger
                                value={EServiceMode.ReplicatedJob}
                                disabled={isAppStopped && savedMode !== EServiceMode.ReplicatedJob}
                            >
                                Replicated Job
                            </TabsTrigger>
                            <TabsTrigger
                                value={EServiceMode.Global}
                                disabled={isAppStopped && savedMode !== EServiceMode.Global}
                            >
                                Global
                            </TabsTrigger>
                            <TabsTrigger
                                value={EServiceMode.GlobalJob}
                                disabled={isAppStopped && savedMode !== EServiceMode.GlobalJob}
                            >
                                Global Job
                            </TabsTrigger>
                        </TabsList>
                    </Tabs>

                    {isAppStopped && (
                        <p className="text-xs text-muted-foreground">
                            Start the app to change its service mode. A stopped app has no running replicas, and modes
                            such as Global cannot express that state.
                        </p>
                    )}
                </div>
            </InfoBlock>

            {isChangingMode && (
                <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                    <span className="font-semibold text-orange-500">Downtime warning:</span> Docker Swarm cannot change
                    the mode of a running service, so saving this will <strong>delete and recreate</strong> the service.
                    The app stops until the new service is up, and its instance history is lost. All other settings are
                    preserved.
                </div>
            )}

            {mode === EServiceMode.Replicated && (
                <InfoBlock
                    titleWidth={220}
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
                        className="max-w-[100px]"
                        min={0}
                    />
                    <FieldError errors={[serviceReplicasError]} />
                </InfoBlock>
            )}

            {mode === EServiceMode.ReplicatedJob && (
                <>
                    <InfoBlock
                        titleWidth={220}
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
                            className="max-w-[100px]"
                            min={0}
                        />
                        <FieldError errors={[jobMaxConcurrentError]} />
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
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
                            className="max-w-[100px]"
                            min={0}
                        />
                        <FieldError errors={[jobTotalCompletionsError]} />
                    </InfoBlock>
                </>
            )}
        </div>
    );
}

interface Props {
    /** Mode currently stored on the service, used to detect a mode switch. */
    savedMode?: EServiceMode;
    /** A stopped app cannot switch mode: the backend rejects it. */
    isAppStopped?: boolean;
}
