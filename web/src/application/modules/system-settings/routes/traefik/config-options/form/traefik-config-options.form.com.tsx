import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox, Field, FieldError, FieldGroup } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useController, useForm, useFormContext } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { TraefikConfigOptions } from "~/system-settings/api/services";

import { ContentBlock, EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import { Textarea } from "@/components/ui/textarea";

import {
    type TraefikConfigOptionsFormInput,
    type TraefikConfigOptionsFormOutput,
    TraefikConfigOptionsFormSchema,
    emptyTraefikConfigOptionsFormDefaults,
} from "../schemas";
import type { TraefikConfigOptionsFormRef } from "../types";

import { mapTraefikConfigOptionsToFormInput } from "./traefik-config-options.form-mappers";

type SchemaInput = TraefikConfigOptionsFormInput;
type SchemaOutput = TraefikConfigOptionsFormOutput;

const LOG_LEVEL_OPTIONS = ["TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC"];

function LogLevelField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "startupCommand.logLevel" });

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Log Level"
                    content="Set the Traefik log level."
                />
            }
        >
            <FieldGroup>
                <Field>
                    <EditableCombobox
                        options={LOG_LEVEL_OPTIONS}
                        value={field.value}
                        onChange={field.onChange}
                        placeholder="Select or type log level"
                        aria-invalid={invalid}
                        className="w-[220px]"
                        inputClassName="w-[220px]"
                        allowClear
                    />
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

function AccessLogField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "startupCommand.accessLog" });

    return (
        <InfoBlock title="Access Log">
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function Http3Field() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "startupCommand.http3" });

    return (
        <InfoBlock title="Enable HTTP3">
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function FastProxyField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "startupCommand.fastProxy" });

    return (
        <InfoBlock title="Fast Proxy (experimental)">
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function ArgsField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "startupCommand.argsText" });

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Args"
                    content="Additional Traefik command-line arguments, one per line."
                />
            }
        >
            <FieldGroup>
                <Field>
                    <Textarea
                        value={field.value}
                        onChange={field.onChange}
                        onBlur={field.onBlur}
                        placeholder="--arg=value"
                        minRows={4}
                        maxRows={10}
                        aria-invalid={invalid}
                    />
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

export function TraefikConfigOptionsForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapTraefikConfigOptionsToFormInput(defaultValues)
            : emptyTraefikConfigOptionsFormDefaults,
        resolver: zodResolver(TraefikConfigOptionsFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapTraefikConfigOptionsToFormInput(defaultValues) : emptyTraefikConfigOptionsFormDefaults,
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
                    <div className={cn(dashedBorderBox, "text-sm leading-6 text-muted-foreground")}>
                        For configuration details, see{" "}
                        <a
                            className="text-blue-500 underline underline-offset-2 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/install-configuration/configuration-options/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            docs
                        </a>
                    </div>

                    <ContentBlock label="Startup Command">
                        <div className="flex flex-col gap-6">
                            <LogLevelField />
                            <AccessLogField />
                            <Http3Field />
                            <FastProxyField />
                            <ArgsField />
                        </div>
                    </ContentBlock>
                </fieldset>

                {children}
            </form>
        </FormProvider>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<TraefikConfigOptionsFormRef>;
    defaultValues?: TraefikConfigOptions;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
