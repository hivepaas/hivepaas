import { useMemo } from "react";

import { FieldError } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Copy, CopyCheck, Grid2x2, Grid2x2Check } from "lucide-react";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components";
import { EServiceMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigAvailabilitySchemaInput, type AppConfigAvailabilitySchemaOutput } from "../schemas";

// A check on the icon means the mode runs to completion rather than staying up.
const MODE_OPTIONS: OptionCard<EServiceMode>[] = [
    {
        value: EServiceMode.Replicated,
        label: "Replicated",
        description: "A set number of instances, kept running",
        icon: Copy,
    },
    {
        value: EServiceMode.ReplicatedJob,
        label: "Replicated Job",
        description: "Runs a set number of times, then stops",
        icon: CopyCheck,
    },
    {
        value: EServiceMode.Global,
        label: "Global",
        description: "One instance on every node, kept running",
        icon: Grid2x2,
    },
    {
        value: EServiceMode.GlobalJob,
        label: "Global Job",
        description: "Runs once on every node, then stops",
        icon: Grid2x2Check,
    },
];

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

    const modeOptions = useMemo(
        () =>
            MODE_OPTIONS.map(option => ({
                ...option,
                disabled: isAppStopped && savedMode !== option.value,
            })),
        [isAppStopped, savedMode],
    );

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
                    <OptionCardGroup
                        options={modeOptions}
                        value={modeField.value}
                        onChange={modeField.onChange}
                        className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 max-w-[800px]"
                    />

                    {isAppStopped && (
                        <p className="text-xs text-muted-foreground">
                            Start the app to change its service mode. A stopped app has no running replicas, and modes
                            such as Global cannot express that state.
                        </p>
                    )}
                </div>
            </InfoBlock>

            {isChangingMode && (
                <div className={cn(dashedBorderBox)}>
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
