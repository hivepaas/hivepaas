import { FieldError } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangle } from "lucide-react";
import { useController, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";
import { z } from "zod";
import type { AppAutoscale, AppAutoscalePausedReason } from "~/projects/api/services";
import { AppAutoscaleCommands } from "~/projects/data";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLink, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Checkbox, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { LOG_HISTORY_UNAVAILABLE_TEXT } from "../../../tabs/logs/building-blocks/log-history-unavailable.constants";

const MAX_REPLICAS_LIMIT = 50;

/** The scale-in delays offered: 1 minute to 1 hour, as the API takes them. */
const SCALE_IN_DELAYS = [
    { value: "1m", label: "1 minute" },
    { value: "2m", label: "2 minutes" },
    { value: "5m", label: "5 minutes" },
    { value: "10m", label: "10 minutes" },
    { value: "15m", label: "15 minutes" },
    { value: "30m", label: "30 minutes" },
    { value: "1h", label: "1 hour" },
];

const AutoscaleSchema = z
    .object({
        enabled: z.boolean(),
        minReplicas: z.number({ message: "Required" }).int().min(1, "At least 1").max(MAX_REPLICAS_LIMIT),
        maxReplicas: z.number({ message: "Required" }).int().min(1, "At least 1").max(MAX_REPLICAS_LIMIT),
        target: z.number({ message: "Required" }).int().min(10, "From 10 %").max(100, "Up to 100 %"),
        scaleInDelay: z.string(),
    })
    .refine(values => values.maxReplicas >= values.minReplicas, {
        path: ["maxReplicas"],
        message: "At least Min Replicas",
    });

type AutoscaleValues = z.infer<typeof AutoscaleSchema>;

function valuesOf(autoscale?: AppAutoscale): AutoscaleValues {
    return {
        enabled: autoscale?.enabled ?? false,
        minReplicas: autoscale?.minReplicas ?? 1,
        maxReplicas: autoscale?.maxReplicas ?? 5,
        target: autoscale?.target ?? 70,
        scaleInDelay: autoscale?.scaleInDelay ?? "5m",
    };
}

/** The reasons an administrator answers in System → Logging. */
const LOGGING_REASONS: AppAutoscalePausedReason[] = ["disabled", "apps-not-collected"];

function pausedText(reason: AppAutoscalePausedReason): string {
    if (reason === "not-replicated") {
        return "Autoscale scales a function that runs a set number of instances: set its Service Mode to Replicated.";
    }
    return `Autoscale reads the function's calls from its stored logs, which cannot be read. ${LOG_HISTORY_UNAVAILABLE_TEXT[reason]}`;
}

/** The replicas the function runs now. A stopped one has none, and autoscale does not start it. */
function replicasText(replicas: number): string {
    if (replicas === 0) {
        return "Stopped: autoscale does not start it";
    }
    return replicas === 1 ? "1 replica now" : `${replicas} replicas now`;
}

/**
 * A function's autoscale: its replicas follow its calls, between Min and Max. Saved on its own, apart from
 * the service's settings below.
 */
export function AutoscaleSection({ projectId, env, appId, autoscale, readOnly = false }: Props) {
    const { control, handleSubmit, reset, watch } = useForm<AutoscaleValues>({
        defaultValues: valuesOf(autoscale),
        resolver: zodResolver(AutoscaleSchema),
        mode: "onSubmit",
    });
    // Only a save moves the settings: a refetch with new replicas or scalings keeps what is being edited.
    useUpdateEffect(() => {
        reset(valuesOf(autoscale));
    }, [autoscale?.updateVer]);

    const { field: enabled } = useController({ control, name: "enabled" });
    const {
        field: minReplicas,
        fieldState: { error: minReplicasError },
    } = useController({ control, name: "minReplicas" });
    const {
        field: maxReplicas,
        fieldState: { error: maxReplicasError },
    } = useController({ control, name: "maxReplicas" });
    const {
        field: target,
        fieldState: { error: targetError },
    } = useController({ control, name: "target" });
    const { field: scaleInDelay } = useController({ control, name: "scaleInDelay" });
    const isEnabled = watch("enabled");

    const savedOn = autoscale?.enabled ?? false;
    const paused = autoscale?.paused ?? null;
    // While its calls cannot be read, autoscale cannot be turned on; one already on says it is paused.
    const cannotTurnOn = Boolean(paused && paused !== "not-replicated" && !savedOn);

    const { mutate: update, isPending } = AppAutoscaleCommands.useUpdateOne({
        onSuccess: response => {
            toast.success("Autoscale updated");
            if (response.data.warning) {
                toast.warning(response.data.warning);
            }
        },
    });

    const delays = SCALE_IN_DELAYS.some(item => item.value === scaleInDelay.value)
        ? SCALE_IN_DELAYS
        : [{ value: scaleInDelay.value, label: scaleInDelay.value }, ...SCALE_IN_DELAYS];

    function onSubmit(values: AutoscaleValues) {
        if (readOnly) {
            return;
        }
        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: { ...values, updateVer: autoscale?.updateVer ?? 0 },
        });
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onSubmit)(event);
            }}
            className="flex flex-col gap-6 px-2"
        >
            <fieldset
                disabled={readOnly}
                className="contents"
            >
                {paused && (savedOn || cannotTurnOn) && (
                    <div className={cn(dashedBorderBox, "flex items-start gap-2")}>
                        <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
                        <span>
                            <span className="font-semibold">
                                {cannotTurnOn ? "Autoscale cannot be turned on." : "Autoscale is paused."}
                            </span>{" "}
                            {pausedText(paused)}{" "}
                            {LOGGING_REASONS.includes(paused) && (
                                <AppLink.Modules
                                    to={ROUTE.systemSettings.logging.configuration.$route}
                                    className="text-link"
                                >
                                    Logging settings
                                </AppLink.Modules>
                            )}
                        </span>
                    </div>
                )}

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Autoscale"
                            content="Scales the function's replicas with its calls, read from its stored logs every 15 seconds: up at once when calls are turned away for Concurrency, down slowly once they have been low for the scale-in delay."
                        />
                    }
                >
                    <div className="flex items-center gap-2">
                        <Checkbox
                            checked={enabled.value}
                            disabled={cannotTurnOn}
                            onCheckedChange={checked => {
                                enabled.onChange(Boolean(checked));
                            }}
                        />
                        {autoscale && (
                            <span className="text-sm text-muted-foreground">{replicasText(autoscale.replicas)}</span>
                        )}
                    </div>
                </InfoBlock>

                {isEnabled && (
                    <>
                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Min Replicas"
                                    content="It never scales below this, however quiet: 1 or more."
                                />
                            }
                        >
                            <InputNumber
                                value={minReplicas.value}
                                onValueChange={value => {
                                    minReplicas.onChange(value);
                                }}
                                className="max-w-[100px]"
                                min={1}
                                max={MAX_REPLICAS_LIMIT}
                            />
                            <FieldError errors={[minReplicasError]} />
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Max Replicas"
                                    content="It never scales above this, however busy: up to 50."
                                />
                            }
                        >
                            <InputNumber
                                value={maxReplicas.value}
                                onValueChange={value => {
                                    maxReplicas.onChange(value);
                                }}
                                className="max-w-[100px]"
                                min={1}
                                max={MAX_REPLICAS_LIMIT}
                            />
                            <FieldError errors={[maxReplicasError]} />
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Target"
                                    content="The share of an instance's Concurrency kept busy. At 70 % and a Concurrency of 16, one instance for every 11 calls in flight."
                                />
                            }
                        >
                            <div className="flex items-center gap-2">
                                <InputNumber
                                    value={target.value}
                                    onValueChange={value => {
                                        target.onChange(value);
                                    }}
                                    className="max-w-[100px]"
                                    min={10}
                                    max={100}
                                />
                                <span className="text-sm text-muted-foreground">%</span>
                            </div>
                            <FieldError errors={[targetError]} />
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Scale-in Delay"
                                    content="How long its calls stay low before it scales in, by half the way to what they need at a time."
                                />
                            }
                        >
                            <Select
                                value={scaleInDelay.value}
                                onValueChange={scaleInDelay.onChange}
                            >
                                <SelectTrigger className="w-full max-w-[200px]">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {delays.map(item => (
                                        <SelectItem
                                            key={item.value}
                                            value={item.value}
                                        >
                                            {item.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </InfoBlock>
                    </>
                )}

                {!readOnly && (
                    <div className="flex justify-end">
                        <ProjectPermissionSubmitButton
                            projectId={projectId}
                            isPending={isPending}
                            label="Save Autoscale"
                            className="min-w-[140px]"
                        />
                    </div>
                )}
            </fieldset>

            {autoscale && autoscale.events.length > 0 && (
                <InfoBlock
                    titleWidth={220}
                    title={<LabelWithInfo label="Latest Scalings" />}
                >
                    <AutoscaleEvents events={autoscale.events} />
                </InfoBlock>
            )}
        </form>
    );
}

function AutoscaleEvents({ events }: { events: AppAutoscale["events"] }) {
    return (
        <div className="max-w-[800px] overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Time</TableHead>
                        <TableHead className="text-right">Replicas</TableHead>
                        <TableHead className="text-right">In flight</TableHead>
                        <TableHead className="text-right">Turned away</TableHead>
                        <TableHead>Why</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {events.map(event => (
                        <TableRow key={`${event.time}-${event.from}-${event.to}`}>
                            <TableCell className="text-xs text-muted-foreground">
                                {new Date(event.time).toLocaleString()}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">
                                {event.from} → {event.to}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{event.inFlight.toFixed(1)}</TableCell>
                            <TableCell className="text-right tabular-nums">{event.throttled}</TableCell>
                            <TableCell className="text-xs">{event.reason}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    autoscale?: AppAutoscale;
    readOnly?: boolean;
}
