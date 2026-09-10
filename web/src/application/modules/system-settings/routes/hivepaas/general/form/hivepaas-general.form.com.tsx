import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangleIcon } from "lucide-react";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { HivePaaSRequestInfoQueries } from "~/system-settings/data";
import type { HivePaaSServiceSettings } from "~/system-settings/domain";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import {
    type HivePaaSKnownProxyProvider,
    PROXY_PROVIDER_IP_URLS,
    PROXY_PROVIDER_OPTIONS,
} from "../hivepaas-general.constants";
import {
    type HivePaaSGeneralFormInput,
    type HivePaaSGeneralFormOutput,
    HivePaaSGeneralFormSchema,
    emptyHivePaaSGeneralFormDefaults,
} from "../schemas";
import type { HivePaaSGeneralFormRef } from "../types";

import { mapHivePaaSServiceSettingsToFormInput } from "./hivepaas-general.form-mappers";

type SchemaInput = HivePaaSGeneralFormInput;
type SchemaOutput = HivePaaSGeneralFormOutput;

function SectionHeader({ children }: PropsWithChildren) {
    return (
        <div className="sticky top-0 z-10 rounded-lg bg-accent px-3 py-2 text-sm font-medium shadow-xs">{children}</div>
    );
}

function NoteBox({ children }: PropsWithChildren) {
    return (
        <div className={cn(dashedBorderBox)}>
            <span className="font-semibold text-orange-500">Note: </span>
            {children}
        </div>
    );
}

function NumberField({ name, label, content, min, max, readOnly }: NumberFieldProps) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name });

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label={label}
                    content={content}
                />
            }
        >
            <FieldGroup>
                <Field>
                    <InputNumber
                        value={field.value}
                        onValueChange={field.onChange}
                        min={min}
                        max={max}
                        decimalScale={0}
                        fixedDecimalScale={false}
                        className="max-w-[110px]"
                        aria-invalid={invalid}
                        disabled={readOnly}
                    />
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

function DurationField({ name, label, content, placeholder }: DurationFieldProps) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name });

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label={label}
                    content={content}
                />
            }
        >
            <FieldGroup>
                <Field>
                    <Input
                        {...field}
                        placeholder={placeholder}
                        className="max-w-[110px]"
                        aria-invalid={invalid}
                    />
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

function RunWorkerInMainAppField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error },
    } = useController({ control, name: "workerSettings.runWorkerInMainApp" });

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Run Worker in Main App"
                    content="Run worker tasks in the main HivePaaS application instance."
                />
            }
        >
            <>
                <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                />
                <FieldError errors={[error]} />
            </>
        </InfoBlock>
    );
}

function ProxyProviderField({ readOnly }: { readOnly: boolean }) {
    const { control, setValue, getValues } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "proxySettings.proxyProvider" });

    const handleProviderChange = (value: string) => {
        field.onChange(value);
        const trimmed = value.trim();
        const currentHops = getValues("proxySettings.proxyHops");
        if (trimmed !== "") {
            if (currentHops === 0) {
                setValue("proxySettings.proxyHops", 2, { shouldValidate: true, shouldDirty: true });
            }
        } else {
            setValue("proxySettings.proxyHops", 0, { shouldValidate: true, shouldDirty: true });
        }
    };

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Proxy Provider"
                    content="Select or enter the reverse proxy provider in front of HivePaaS to help configure trusted client IPs."
                />
            }
        >
            <FieldGroup>
                <Field>
                    <EditableCombobox
                        options={PROXY_PROVIDER_OPTIONS}
                        value={field.value}
                        onChange={handleProviderChange}
                        placeholder="select provider"
                        aria-invalid={invalid}
                        className="max-w-[280px]"
                        inputClassName="max-w-[280px]"
                        disabled={readOnly}
                    />
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

function TrustedIPsField({ readOnly }: { readOnly: boolean }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "proxySettings.trustedIPsText" });
    const proxyProvider = useWatch({ control, name: "proxySettings.proxyProvider" });
    const showIPsUrl = PROXY_PROVIDER_IP_URLS[proxyProvider as HivePaaSKnownProxyProvider];

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Trusted IPs"
                    content="IP addresses or CIDR ranges trusted to set forwarded headers. One entry per line."
                />
            }
        >
            <div className="flex w-full max-w-[560px] items-start gap-3">
                <div className="min-w-0 flex-1">
                    <Textarea
                        {...field}
                        onChange={field.onChange}
                        placeholder={"1.2.3.4\n2001:0DC8:1005:2F43:0BCD:FFFF"}
                        className="min-h-[150px] h-[180px] resize-y"
                        aria-invalid={invalid}
                        disabled={readOnly}
                    />
                    <FieldError errors={[error]} />
                </div>
                {showIPsUrl ? (
                    <a
                        className="shrink-0 pt-2 text-sm text-blue-500 hover:text-blue-600"
                        href={showIPsUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        Show IPs
                    </a>
                ) : null}
            </div>
        </InfoBlock>
    );
}

/**
 * What the browser's own request says about the proxies in front.
 *
 * The hop count cannot be worked out from the deployment: whether a proxy adds
 * itself to X-Forwarded-For is that proxy's decision. It has to be measured, and
 * the request that painted this page took exactly the path real traffic takes -
 * so the measurement is right here for the asking, and guessing is never the
 * better option.
 *
 * It sits outside the "a provider is configured" branch on purpose. The most
 * damaging way to get this section wrong is not a hop count that is off by one:
 * it is a proxy that is really there and was never declared, because then traefik
 * trusts nobody, every allowlist compares against the proxy's address instead of
 * the caller's, and nothing anywhere says so. That case only shows up when the
 * measurement is visible before a provider has been chosen.
 */
function MeasuredProxyTopology({ readOnly, hasProxyProvider }: { readOnly: boolean; hasProxyProvider: boolean }) {
    const { control, setValue } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const currentHops = useWatch({ control, name: "proxySettings.proxyHops" });
    const requestInfo = HivePaaSRequestInfoQueries.useFindOne();

    const info = requestInfo.data?.data;
    if (info == null) {
        // Silent when it cannot be read. It is a hint beside a field, and an error
        // about a missing hint helps nobody who is trying to fill the field in.
        return null;
    }

    const detectedProxy = info.suggestedProxyHops > 0;
    const undeclaredProxy = detectedProxy && !hasProxyProvider;
    // Offering to fill in the hops without a provider would put the form into a
    // state its own schema rejects - hops above zero require one.
    const canApply = hasProxyProvider && !readOnly && info.suggestedProxyHops !== currentHops;

    return (
        <div className="flex flex-col gap-1.5 rounded-lg border bg-background/50 p-3 text-xs">
            {undeclaredProxy ? (
                <p className="flex items-center gap-2 font-semibold text-amber-600 dark:text-amber-400">
                    <AlertTriangleIcon className="size-3.5 shrink-0" />
                    This request arrived through {info.suggestedProxyHops} proxy hop
                    {info.suggestedProxyHops === 1 ? "" : "s"}, but no proxy provider is configured
                </p>
            ) : (
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-muted-foreground">
                        {detectedProxy
                            ? "Measured from this browser's request:"
                            : "No proxy detected in front of HivePaaS."}
                    </span>
                    {detectedProxy && <span className="font-mono font-semibold">{info.suggestedProxyHops}</span>}
                    {hasProxyProvider &&
                        (canApply ? (
                            <button
                                type="button"
                                className="text-link underline-offset-4 hover:underline"
                                onClick={() => {
                                    setValue("proxySettings.proxyHops", info.suggestedProxyHops, {
                                        shouldValidate: true,
                                        shouldDirty: true,
                                    });
                                }}
                            >
                                Use {info.suggestedProxyHops}
                            </button>
                        ) : (
                            <span className="text-emerald-600 dark:text-emerald-400">matches the value below</span>
                        ))}
                </div>
            )}

            {info.forwardedFor.length > 0 && (
                <div className="text-muted-foreground">
                    <span>X-Forwarded-For: </span>
                    <span className="font-mono break-all">{info.forwardedFor.join(" → ")}</span>
                </div>
            )}

            {/*
             * The check that makes the number trustworthy, and the reason the
             * chain above is shown at all: if the first entry is not the
             * operator's own public address, the reading is of some other path
             * and the suggestion is wrong.
             */}
            <p className="text-muted-foreground">{info.explanation}</p>
        </div>
    );
}

function ProxyHopsField({ readOnly }: { readOnly: boolean }) {
    return (
        <NumberField
            name="proxySettings.proxyHops"
            label="Proxy Hops"
            content="Number of reverse proxy hops in front of HivePaaS to resolve the client IP address."
            min={0}
            max={10}
            readOnly={readOnly}
        />
    );
}

function ProxyConfigurationSection({ readOnly }: { readOnly: boolean }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const proxyProvider = useWatch({ control, name: "proxySettings.proxyProvider" });
    const hasProxyProvider = proxyProvider.trim() !== "";

    return (
        <>
            <SectionHeader>Proxy Configuration</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <MeasuredProxyTopology
                    readOnly={readOnly}
                    hasProxyProvider={hasProxyProvider}
                />
                <ProxyProviderField readOnly={readOnly} />
                {hasProxyProvider ? (
                    <>
                        <TrustedIPsField readOnly={readOnly} />
                        <ProxyHopsField readOnly={readOnly} />
                    </>
                ) : null}
            </div>
        </>
    );
}

export function HivePaaSGeneralForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapHivePaaSServiceSettingsToFormInput(defaultValues)
            : emptyHivePaaSGeneralFormDefaults,
        resolver: zodResolver(HivePaaSGeneralFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapHivePaaSServiceSettingsToFormInput(defaultValues) : emptyHivePaaSGeneralFormDefaults,
        );
    }, [defaultValues]);

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<SchemaInput>) => {
                methods.reset({
                    ...methods.getValues(),
                    ...values,
                } as SchemaInput);
            },
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }

                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as FieldPath<SchemaInput>,
                        { message, type: "manual" },
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods],
    );

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void methods.handleSubmit(onSubmit)(event);
                }}
                className="flex flex-col gap-6"
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
                    <NoteBox>
                        By default, HivePaaS runs a single instance for both the main application and the worker. This
                        model is the most resource-efficient. If you need a system with higher processing capacity, you
                        can increase the number of instances and run the worker separately from the main application.
                    </NoteBox>

                    <SectionHeader>Service Configuration</SectionHeader>
                    <div className="flex flex-col gap-6 px-3">
                        <NumberField
                            name="appSettings.replicas"
                            label="Replicas"
                            content="Number of HivePaaS main application replicas."
                            min={1}
                            max={100}
                        />
                    </div>

                    <SectionHeader>Worker Configuration</SectionHeader>
                    <div className="flex flex-col gap-6 px-3">
                        <NumberField
                            name="workerSettings.replicas"
                            label="Replicas"
                            content="Number of HivePaaS worker replicas."
                            min={0}
                            max={100}
                        />
                        <NumberField
                            name="workerSettings.concurrency"
                            label="Concurrency"
                            content="Maximum worker task concurrency."
                            min={1}
                            max={100}
                        />
                        <RunWorkerInMainAppField />
                    </div>

                    <SectionHeader>Task Queue Configuration</SectionHeader>
                    <div className="flex flex-col gap-6 px-3">
                        <DurationField
                            name="taskSettings.taskCheckInterval"
                            label="Task Check Interval"
                            content="How often HivePaaS checks queued tasks."
                            placeholder="10m"
                        />
                        <DurationField
                            name="taskSettings.taskCreateInterval"
                            label="Task Creation Interval"
                            content="How often HivePaaS creates queued tasks."
                            placeholder="10m"
                        />
                    </div>

                    <SectionHeader>Periodic Job Configuration</SectionHeader>
                    <div className="flex flex-col gap-6 px-3">
                        <NoteBox>
                            If the system has many periodic jobs, you can decrease the Base Interval to evenly
                            distribute the system load. Note: The HivePaaS app will restart, which may interrupt ongoing
                            tasks.
                        </NoteBox>
                        <DurationField
                            name="periodicSettings.baseInterval"
                            label="Base Interval"
                            content="Base interval for HivePaaS periodic jobs."
                            placeholder="15s"
                        />
                        <NumberField
                            name="periodicSettings.batchSize"
                            label="Batch Size"
                            content="Maximum number of due periodic jobs to fetch and process in a single tick (per second)"
                            min={1}
                            max={10000}
                        />
                    </div>

                    <ProxyConfigurationSection readOnly={readOnly} />
                </fieldset>

                {children}
            </form>
        </FormProvider>
    );
}

type NumberFieldPath = Extract<
    FieldPath<SchemaInput>,
    | "appSettings.replicas"
    | "workerSettings.replicas"
    | "workerSettings.concurrency"
    | "periodicSettings.batchSize"
    | "proxySettings.proxyHops"
>;

type DurationFieldPath = Extract<
    FieldPath<SchemaInput>,
    "taskSettings.taskCheckInterval" | "taskSettings.taskCreateInterval" | "periodicSettings.baseInterval"
>;

type NumberFieldProps = {
    name: NumberFieldPath;
    label: string;
    content: string;
    min: number;
    max: number;
    readOnly?: boolean;
};

type DurationFieldProps = {
    name: DurationFieldPath;
    label: string;
    content: string;
    placeholder: string;
};

type Props = PropsWithChildren<{
    ref?: React.Ref<HivePaaSGeneralFormRef>;
    defaultValues?: HivePaaSServiceSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
