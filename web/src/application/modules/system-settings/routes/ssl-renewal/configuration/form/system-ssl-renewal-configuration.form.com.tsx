import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import type { SystemSslRenewalSettings } from "~/system-settings/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { EnabledConfigurationFields, EnabledField } from "../building-blocks";
import {
    type SystemSslRenewalConfigurationFormInput,
    type SystemSslRenewalConfigurationFormOutput,
    SystemSslRenewalConfigurationFormSchema,
} from "../schemas";
import type { SystemSslRenewalConfigurationFormRef } from "../types";

import {
    emptySystemSslRenewalConfigurationFormDefaults,
    mapSystemSslRenewalSettingsToFormInput,
} from "./system-ssl-renewal-configuration.form-mappers";

type SchemaInput = SystemSslRenewalConfigurationFormInput;
type SchemaOutput = SystemSslRenewalConfigurationFormOutput;

function useSystemSslRenewalFormMethods(defaultValues?: SystemSslRenewalSettings) {
    return useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSystemSslRenewalSettingsToFormInput(defaultValues)
            : emptySystemSslRenewalConfigurationFormDefaults,
        resolver: zodResolver(SystemSslRenewalConfigurationFormSchema),
        mode: "onSubmit",
    });
}

export function SystemSslRenewalConfigurationForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const methods = useSystemSslRenewalFormMethods(defaultValues);

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
    ref?: React.Ref<SystemSslRenewalConfigurationFormRef>;
    defaultValues?: SystemSslRenewalSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
