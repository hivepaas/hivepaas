import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppNetworkSettings } from "~/projects/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { DNSFields, EndpointPortConfigFields, HostsFileEntriesFields, NetworksFields } from "../building-blocks";
import {
    AppConfigNetworksFormSchema,
    type AppConfigNetworksFormSchemaInput,
    type AppConfigNetworksFormSchemaOutput,
    emptyAppConfigNetworksFormDefaults,
} from "../schemas";
import { type AppConfigNetworksFormRef } from "../types";

import { mapAppNetworkSettingsToFormInput } from "./app-config-networks.form-mappers";

type SchemaInput = AppConfigNetworksFormSchemaInput;
type SchemaOutput = AppConfigNetworksFormSchemaOutput;

export function AppConfigNetworksForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppNetworkSettingsToFormInput(defaultValues)
            : emptyAppConfigNetworksFormDefaults,
        resolver: zodResolver(AppConfigNetworksFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapAppNetworkSettingsToFormInput(defaultValues) : emptyAppConfigNetworksFormDefaults,
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
                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Network Settings</h3>
                    <div className="flex flex-col gap-6 px-2">
                        <NetworksFields />
                        <HostsFileEntriesFields />
                        <DNSFields />
                    </div>
                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Endpoint & Port Config</h3>
                    <div className="flex flex-col gap-6 px-2">
                        <EndpointPortConfigFields />
                    </div>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigNetworksFormRef>;
    defaultValues?: AppNetworkSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
