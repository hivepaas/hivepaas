import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
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
    return <div className="rounded-lg bg-muted px-4 py-3 text-sm font-semibold text-foreground">{children}</div>;
}

function NoteBox({ children }: PropsWithChildren) {
    return (
        <div className={cn(dashedBorderBox, "text-sm leading-6")}>
            <span className="font-semibold text-orange-500">Note: </span>
            {children}
        </div>
    );
}

function NumberField({ name, label, content, min, max }: NumberFieldProps) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name });

    return (
        <InfoBlock
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
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "proxySettings.proxyProvider" });

    return (
        <InfoBlock
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
                        onChange={field.onChange}
                        placeholder="select provider"
                        aria-invalid={invalid}
                        className="max-w-[280px]"
                        inputClassName="max-w-[280px]"
                        allowClear
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

function ProxyConfigurationSection({ readOnly }: { readOnly: boolean }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const proxyProvider = useWatch({ control, name: "proxySettings.proxyProvider" });
    const hasProxyProvider = proxyProvider.trim() !== "";

    return (
        <>
            <SectionHeader>Proxy Configuration</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <ProxyProviderField readOnly={readOnly} />
                {hasProxyProvider ? <TrustedIPsField readOnly={readOnly} /> : null}
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
    "appSettings.replicas" | "workerSettings.replicas" | "workerSettings.concurrency" | "periodicSettings.batchSize"
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
