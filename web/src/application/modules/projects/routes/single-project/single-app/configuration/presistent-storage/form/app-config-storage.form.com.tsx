import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppStorageSettings } from "~/projects/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { StorageMountsFields } from "../building-blocks";
import {
    AppConfigStorageFormSchema,
    type AppConfigStorageFormSchemaInput,
    type AppConfigStorageFormSchemaOutput,
    emptyAppConfigStorageFormDefaults,
} from "../schemas";
import { type AppConfigStorageFormRef } from "../types";

import { mapAppStorageSettingsToFormInput } from "./app-config-storage.form-mappers";

type SchemaInput = AppConfigStorageFormSchemaInput;
type SchemaOutput = AppConfigStorageFormSchemaOutput;

export function AppConfigStorageForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppStorageSettingsToFormInput(defaultValues)
            : emptyAppConfigStorageFormDefaults,
        resolver: zodResolver(AppConfigStorageFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapAppStorageSettingsToFormInput(defaultValues) : emptyAppConfigStorageFormDefaults,
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
                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Persistent Storage</h3>
                    <div className="flex flex-col gap-6 px-2">
                        <StorageMountsFields />
                    </div>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigStorageFormRef>;
    defaultValues?: AppStorageSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
