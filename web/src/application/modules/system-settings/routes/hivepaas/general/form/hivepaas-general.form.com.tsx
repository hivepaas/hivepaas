import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useController, useForm, useFormContext } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { HivePaaSServiceSettings } from "~/system-settings/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

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
        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
            <span className="text-orange-500">Note: </span>
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

                    <SectionHeader>Health Check Configuration</SectionHeader>
                    <div className="flex flex-col gap-6 px-3">
                        <DurationField
                            name="healthcheckSettings.baseInterval"
                            label="Base Interval"
                            content="Base interval for HivePaaS health checks."
                            placeholder="15s"
                        />
                    </div>
                </fieldset>

                {children}
            </form>
        </FormProvider>
    );
}

type NumberFieldPath = Extract<
    FieldPath<SchemaInput>,
    "appSettings.replicas" | "workerSettings.replicas" | "workerSettings.concurrency"
>;

type DurationFieldPath = Extract<
    FieldPath<SchemaInput>,
    "taskSettings.taskCheckInterval" | "taskSettings.taskCreateInterval" | "healthcheckSettings.baseInterval"
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
