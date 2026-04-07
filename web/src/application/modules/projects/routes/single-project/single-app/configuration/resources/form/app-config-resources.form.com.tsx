import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppResourceSettings } from "~/projects/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { CapabilitiesFields, ResourceLimitFields, ResourceReservationFields, UlimitsFields } from "../building-blocks";
import {
    AppConfigResourcesFormSchema,
    type AppConfigResourcesFormSchemaInput,
    type AppConfigResourcesFormSchemaOutput,
    emptyAppConfigResourcesFormDefaults,
} from "../schemas";
import { type AppConfigResourcesFormRef } from "../types";

import { mapAppResourceSettingsToFormInput } from "./app-config-resources.form-mappers";

type SchemaInput = AppConfigResourcesFormSchemaInput;
type SchemaOutput = AppConfigResourcesFormSchemaOutput;

export function AppConfigResourcesForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppResourceSettingsToFormInput(defaultValues)
            : emptyAppConfigResourcesFormDefaults,
        resolver: zodResolver(AppConfigResourcesFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapAppResourceSettingsToFormInput(defaultValues) : emptyAppConfigResourcesFormDefaults,
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
                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Resource Reservation</h3>
                    <ResourceReservationFields />

                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Resource Limit</h3>
                    <ResourceLimitFields />

                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Ulimits</h3>
                    <UlimitsFields />

                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Capabilities</h3>
                    <CapabilitiesFields />

                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigResourcesFormRef>;
    defaultValues?: AppResourceSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
