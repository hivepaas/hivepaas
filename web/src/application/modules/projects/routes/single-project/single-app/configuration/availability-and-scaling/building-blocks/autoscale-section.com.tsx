import type { PropsWithChildren } from "react";

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
import type { AppAutoscale } from "~/projects/api/services";
import { AppAutoscaleCommands } from "~/projects/data";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLink, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Checkbox, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";

import { AutoscaleEvents } from "./autoscale-events.com";
import { AutoscaleTestGuide } from "./autoscale-test-guide.com";
import { LOGGING_REASONS, SCALE_IN_DELAYS, pausedText, replicasText, signalText } from "./autoscale-texts";

const MAX_REPLICAS_LIMIT = 50;
/** What Requests Per Instance starts at when it is turned on. */
const REQUESTS_TARGET_DEFAULT = 10;
const CPU_TARGET_DEFAULT = 70;

const AutoscaleSchema = z
    .object({
        isFunction: z.boolean(),
        enabled: z.boolean(),
        minReplicas: z.number({ message: "Required" }).int().min(1, "At least 1").max(MAX_REPLICAS_LIMIT),
        maxReplicas: z.number({ message: "Required" }).int().min(1, "At least 1").max(MAX_REPLICAS_LIMIT),
        target: z.number({ message: "Required" }).int(),
        requestsOn: z.boolean(),
        requestsTarget: z.number({ message: "Required" }).int(),
        cpuOn: z.boolean(),
        cpuTarget: z.number({ message: "Required" }).int(),
        scaleInDelay: z.string(),
    })
    .superRefine((values, ctx) => {
        const outside = (value: number, min: number, max: number) => value < min || value > max;
        if (values.maxReplicas < values.minReplicas) {
            ctx.addIssue({ code: "custom", path: ["maxReplicas"], message: "At least Min Replicas" });
        }
        if (values.isFunction) {
            if (outside(values.target, 10, 100)) {
                ctx.addIssue({ code: "custom", path: ["target"], message: "From 10 to 100 %" });
            }
            return;
        }
        if (values.enabled && !values.requestsOn && !values.cpuOn) {
            ctx.addIssue({ code: "custom", path: ["requestsOn"], message: "Scale on requests, CPU or both" });
        }
        if (values.requestsOn && outside(values.requestsTarget, 1, 1000)) {
            ctx.addIssue({ code: "custom", path: ["requestsTarget"], message: "From 1 to 1000" });
        }
        if (values.cpuOn && outside(values.cpuTarget, 10, 100)) {
            ctx.addIssue({ code: "custom", path: ["cpuTarget"], message: "From 10 to 100 %" });
        }
    });

type AutoscaleValues = z.infer<typeof AutoscaleSchema>;

function valuesOf(autoscale?: AppAutoscale): AutoscaleValues {
    const requestsTarget = autoscale?.requestsTarget ?? 0;
    const cpuTarget = autoscale?.cpuTarget ?? CPU_TARGET_DEFAULT;
    return {
        isFunction: autoscale?.isFunction ?? false,
        enabled: autoscale?.enabled ?? false,
        minReplicas: autoscale?.minReplicas ?? 1,
        maxReplicas: autoscale?.maxReplicas ?? 5,
        target: autoscale?.target ?? 70,
        requestsOn: requestsTarget > 0,
        requestsTarget: requestsTarget > 0 ? requestsTarget : REQUESTS_TARGET_DEFAULT,
        cpuOn: cpuTarget > 0,
        cpuTarget: cpuTarget > 0 ? cpuTarget : CPU_TARGET_DEFAULT,
        scaleInDelay: autoscale?.scaleInDelay ?? "5m",
    };
}

/**
 * An app's autoscale: its replicas follow its load, between Min and Max - a function's its calls, any other app's
 * its requests, its CPU or both. Saved on its own, apart from the service's settings below.
 */
export function AutoscaleSection({ projectId, env, appId, autoscale, readOnly = false }: Props) {
    const { control, handleSubmit, reset, watch, setError } = useForm<AutoscaleValues>({
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
    const {
        field: requestsOn,
        fieldState: { error: requestsOnError },
    } = useController({ control, name: "requestsOn" });
    const {
        field: requestsTarget,
        fieldState: { error: requestsTargetError },
    } = useController({ control, name: "requestsTarget" });
    const { field: cpuOn } = useController({ control, name: "cpuOn" });
    const {
        field: cpuTarget,
        fieldState: { error: cpuTargetError },
    } = useController({ control, name: "cpuTarget" });
    const { field: scaleInDelay } = useController({ control, name: "scaleInDelay" });
    const isEnabled = watch("enabled");
    const [watchedRequestsOn, watchedRequestsTarget, watchedMaxReplicas] = watch([
        "requestsOn",
        "requestsTarget",
        "maxReplicas",
    ]);

    const isFunction = autoscale?.isFunction ?? false;
    const savedOn = autoscale?.enabled ?? false;
    const paused = autoscale?.paused ?? null;
    // A function whose calls cannot be read, or an app publishing a port on its node, cannot have it turned on. An
    // app's unreadable signals are said by each, as which it scales on is being chosen.
    const cannotTurnOn =
        !savedOn && paused !== null && (isFunction ? paused !== "not-replicated" : paused === "host-ports");

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
        // Turned on, an app must read one of what it scales on: the API refuses it otherwise. One already on is
        // saved, and says it is paused.
        const readsOne =
            (values.requestsOn && !autoscale?.requestsUnavailable) || (values.cpuOn && !autoscale?.cpuUnavailable);
        if (!values.isFunction && values.enabled && !savedOn && !readsOne) {
            setError("requestsOn", {
                type: "manual",
                message: "What it scales on cannot be read now, as said above: autoscale can be turned on once it can.",
            });
            return;
        }
        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: {
                enabled: values.enabled,
                minReplicas: values.minReplicas,
                maxReplicas: values.maxReplicas,
                target: values.target,
                requestsTarget: !values.isFunction && values.requestsOn ? values.requestsTarget : 0,
                cpuTarget: !values.isFunction && values.cpuOn ? values.cpuTarget : 0,
                scaleInDelay: values.scaleInDelay,
                updateVer: autoscale?.updateVer ?? 0,
            },
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
                    <Warning>
                        <span className="font-semibold">
                            {cannotTurnOn ? "Autoscale cannot be turned on." : "Autoscale is paused."}
                        </span>{" "}
                        {pausedText(paused, isFunction)} {LOGGING_REASONS.includes(paused) && <LoggingLink />}
                    </Warning>
                )}

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Autoscale"
                            content={
                                isFunction
                                    ? "Scales the function's replicas with its calls, read from its stored logs every 15 seconds: up at once when calls are turned away for Concurrency or come in a burst, else once they have needed more for 30 seconds; down slowly once they have been low for the scale-in delay."
                                    : "Scales the app's replicas with its requests, its CPU or both, read every 15 seconds: up at once for a burst of requests, else once they have needed more for 30 seconds; down slowly once they have been low for the scale-in delay."
                            }
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

                        {isFunction ? (
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
                        ) : (
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Scale On"
                                        content="What its replicas follow: its requests, its CPU or both. With both, the one asking for more instances wins. A request counts when it ends, for a minute at most: an app serving mostly WebSockets, server-sent events or long polls scales better on CPU."
                                    />
                                }
                            >
                                <div className="flex flex-col gap-4">
                                    <Signal
                                        label="Requests"
                                        checked={requestsOn.value}
                                        onCheckedChange={requestsOn.onChange}
                                        unavailable={autoscale?.requestsUnavailable ?? null}
                                    >
                                        <InputNumber
                                            value={requestsTarget.value}
                                            onValueChange={value => {
                                                requestsTarget.onChange(value);
                                            }}
                                            className="max-w-[100px]"
                                            min={1}
                                            max={1000}
                                            disabled={!requestsOn.value}
                                        />
                                        <span className="text-sm text-muted-foreground">
                                            in flight per instance, through its domains
                                        </span>
                                    </Signal>
                                    <FieldError errors={[requestsTargetError]} />

                                    <Signal
                                        label="CPU"
                                        checked={cpuOn.value}
                                        onCheckedChange={cpuOn.onChange}
                                        unavailable={autoscale?.cpuUnavailable ?? null}
                                    >
                                        <InputNumber
                                            value={cpuTarget.value}
                                            onValueChange={value => {
                                                cpuTarget.onChange(value);
                                            }}
                                            className="max-w-[100px]"
                                            min={10}
                                            max={100}
                                            disabled={!cpuOn.value}
                                        />
                                        <span className="text-sm text-muted-foreground">
                                            % of an instance&rsquo;s CPU limit, or its reservation
                                        </span>
                                    </Signal>
                                    <FieldError errors={[cpuTargetError, requestsOnError]} />
                                </div>
                            </InfoBlock>
                        )}

                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Scale-in Delay"
                                    content="How long its load stays low before it scales in, by half the way to what it needs at a time."
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

                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Test It"
                                    content="Load it, and watch it scale: the steps, and a command to load it with."
                                />
                            }
                        >
                            <AutoscaleTestGuide
                                projectId={projectId}
                                env={env}
                                appId={appId}
                                isFunction={isFunction}
                                requestsTarget={watchedRequestsOn ? watchedRequestsTarget : REQUESTS_TARGET_DEFAULT}
                                maxReplicas={watchedMaxReplicas}
                            />
                        </InfoBlock>

                        {!isFunction && autoscale?.writableMounts && (
                            <Warning>
                                The app writes to a volume or a bind mount. Its replicas on one node share it, and those
                                on another node each have their own: an app that keeps its data there, such as a
                                database, must not autoscale.
                            </Warning>
                        )}
                    </>
                )}

                {savedOn && autoscale && autoscale.pending > 0 && (
                    <Warning>
                        {autoscale.pending === 1 ? "A replica" : `${autoscale.pending} replicas`} cannot start: the
                        cluster may have no room for them. Autoscale scales it no further out until they run; see its{" "}
                        <AppLink.Basic
                            to={ROUTE.projects.single.apps.single.instances.$route(projectId, env, appId)}
                            className="text-link"
                        >
                            Instances
                        </AppLink.Basic>
                        .
                    </Warning>
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
                    <AutoscaleEvents
                        events={autoscale.events}
                        isFunction={isFunction}
                    />
                </InfoBlock>
            )}
        </form>
    );
}

/** One signal an app can scale on: whether it does, its target, and why it cannot be read now. */
function Signal({ label, checked, onCheckedChange, unavailable, children }: PropsWithChildren<SignalProps>) {
    return (
        <div className="flex flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
                <Checkbox
                    checked={checked}
                    onCheckedChange={value => {
                        onCheckedChange(Boolean(value));
                    }}
                />
                <span className="w-[72px] text-sm font-medium">{label}</span>
                {children}
            </div>
            {unavailable && (
                <p className="text-xs text-muted-foreground">
                    <AlertTriangle className="mr-1 inline size-3 text-amber-600 dark:text-amber-500" />
                    {signalText(unavailable)} {LOGGING_REASONS.includes(unavailable) && <LoggingLink />}
                </p>
            )}
        </div>
    );
}

function Warning({ children }: PropsWithChildren) {
    return (
        <div className={cn(dashedBorderBox, "flex items-start gap-2")}>
            <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
            <span>{children}</span>
        </div>
    );
}

function LoggingLink() {
    return (
        <AppLink.Modules
            to={ROUTE.systemSettings.logging.configuration.$route}
            className="text-link"
        >
            Logging settings
        </AppLink.Modules>
    );
}

interface SignalProps {
    label: string;
    checked: boolean;
    onCheckedChange: (checked: boolean) => void;
    /** Why it cannot be read now; null when it can. */
    unavailable: string | null;
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    autoscale?: AppAutoscale;
    readOnly?: boolean;
}
