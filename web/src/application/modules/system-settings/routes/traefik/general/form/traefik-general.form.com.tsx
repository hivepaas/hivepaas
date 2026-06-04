import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useController, useForm, useFormContext } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { TraefikServiceSettings } from "~/system-settings/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import {
    type TraefikGeneralFormInput,
    type TraefikGeneralFormOutput,
    TraefikGeneralFormSchema,
    emptyTraefikGeneralFormDefaults,
} from "../schemas";
import type { TraefikGeneralFormRef } from "../types";

import { mapTraefikServiceSettingsToFormInput } from "./traefik-general.form-mappers";

type SchemaInput = TraefikGeneralFormInput;
type SchemaOutput = TraefikGeneralFormOutput;

function SectionHeader({ children }: PropsWithChildren) {
    return <div className="rounded-lg bg-muted px-4 py-3 text-sm font-semibold text-foreground">{children}</div>;
}

function ReplicasField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "appSettings.replicas" });

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Replicas"
                    content="Number of Traefik service replicas."
                />
            }
        >
            <FieldGroup>
                <Field>
                    <InputNumber
                        value={field.value}
                        onValueChange={field.onChange}
                        min={1}
                        max={100}
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

export function TraefikGeneralForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapTraefikServiceSettingsToFormInput(defaultValues)
            : emptyTraefikGeneralFormDefaults,
        resolver: zodResolver(TraefikGeneralFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapTraefikServiceSettingsToFormInput(defaultValues) : emptyTraefikGeneralFormDefaults,
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
                    <SectionHeader>Service Configuration</SectionHeader>
                    <div className="flex flex-col gap-6 px-3">
                        <ReplicasField />
                    </div>
                </fieldset>

                {children}
            </form>
        </FormProvider>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<TraefikGeneralFormRef>;
    defaultValues?: TraefikServiceSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
