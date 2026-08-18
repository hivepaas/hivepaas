import React, { useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppPlacementSettings } from "~/settings/domain";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { AppPlacementFields } from "../building-blocks";
import {
    SettingsAppPlacementFormSchema,
    type SettingsAppPlacementFormSchemaInput,
    type SettingsAppPlacementFormSchemaOutput,
    emptySettingsAppPlacementFormDefaults,
} from "../schemas";
import type { SettingsAppPlacementFormRef } from "../types";

import { mapSettingsAppPlacementToFormInput } from "./settings-app-placement.form-mappers";

type SchemaInput = SettingsAppPlacementFormSchemaInput;
type SchemaOutput = SettingsAppPlacementFormSchemaOutput;

export function SettingsAppPlacementForm({ ref, defaultValues, onSubmit, readOnly = false, note, footer }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSettingsAppPlacementToFormInput(defaultValues)
            : emptySettingsAppPlacementFormDefaults,
        resolver: zodResolver(SettingsAppPlacementFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapSettingsAppPlacementToFormInput(defaultValues) : emptySettingsAppPlacementFormDefaults,
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
                <ContentBlock label="App Placement Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            {note}
                            <AppPlacementFields />
                        </div>
                    </fieldset>
                </ContentBlock>

                {footer}
            </form>
        </FormProvider>
    );
}

type Props = {
    ref?: React.Ref<SettingsAppPlacementFormRef>;
    defaultValues?: AppPlacementSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
    note?: React.ReactNode;
    footer: React.ReactNode;
};
