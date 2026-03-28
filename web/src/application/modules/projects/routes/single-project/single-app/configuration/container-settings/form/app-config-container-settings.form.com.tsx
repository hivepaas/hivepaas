import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { type AppContainerSettings } from "~/projects/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { GeneralFields, LabelsFields, RestartPolicyFields, SecurityFields } from "../building-blocks";
import {
    AppConfigContainerSettingsFormSchema,
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
    emptyAppConfigContainerSettingsFormDefaults,
} from "../schemas";
import { type AppConfigContainerSettingsFormRef } from "../types";

import { mapAppContainerSettingsToFormInput } from "./app-config-container-settings.form-mappers";

type SchemaInput = AppConfigContainerSettingsFormSchemaInput;
type SchemaOutput = AppConfigContainerSettingsFormSchemaOutput;

export function AppConfigContainerSettingsForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppContainerSettingsToFormInput(defaultValues)
            : emptyAppConfigContainerSettingsFormDefaults,
        resolver: zodResolver(AppConfigContainerSettingsFormSchema),
        mode: "onSubmit",
    });

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
                        {
                            shouldFocus: index === 0,
                        },
                    );
                });
            },
        }),
        [methods],
    );

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <GeneralFields />
                    <LabelsFields />
                    <RestartPolicyFields />
                    <SecurityFields />
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigContainerSettingsFormRef>;
    defaultValues?: AppContainerSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
