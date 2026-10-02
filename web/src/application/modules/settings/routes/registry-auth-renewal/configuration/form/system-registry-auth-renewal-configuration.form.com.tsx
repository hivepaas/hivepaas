import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import type { SystemRegistryAuthRenewalSettings } from "~/system-settings/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { EnabledConfigurationFields, EnabledField } from "../building-blocks";
import {
    type SystemRegistryAuthRenewalConfigurationFormInput,
    type SystemRegistryAuthRenewalConfigurationFormOutput,
    SystemRegistryAuthRenewalConfigurationFormSchema,
} from "../schemas";
import type { SystemRegistryAuthRenewalConfigurationFormRef } from "../types";

import {
    emptySystemRegistryAuthRenewalConfigurationFormDefaults,
    mapSystemRegistryAuthRenewalSettingsToFormInput,
} from "./system-registry-auth-renewal-configuration.form-mappers";

type SchemaInput = SystemRegistryAuthRenewalConfigurationFormInput;
type SchemaOutput = SystemRegistryAuthRenewalConfigurationFormOutput;

function useSystemRegistryAuthRenewalFormMethods(defaultValues?: SystemRegistryAuthRenewalSettings) {
    return useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSystemRegistryAuthRenewalSettingsToFormInput(defaultValues)
            : emptySystemRegistryAuthRenewalConfigurationFormDefaults,
        resolver: zodResolver(SystemRegistryAuthRenewalConfigurationFormSchema),
        mode: "onSubmit",
    });
}

export function SystemRegistryAuthRenewalConfigurationForm({
    ref,
    defaultValues,
    onSubmit,
    readOnly = false,
    children,
}: Props) {
    const methods = useSystemRegistryAuthRenewalFormMethods(defaultValues);

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
                        if (readOnly) {
                            return;
                        }

                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
                    >
                        <EnabledField />
                        <EnabledConfigurationFields
                            nextRuns={defaultValues?.nextRuns ?? []}
                            readOnly={readOnly}
                        />
                    </fieldset>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SystemRegistryAuthRenewalConfigurationFormRef>;
    defaultValues?: SystemRegistryAuthRenewalSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
