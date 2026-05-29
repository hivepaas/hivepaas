import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppNetworkSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

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

export function AppConfigNetworksForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
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
                        <ContentBlock label="Network Settings">
                            <div className="flex flex-col gap-6">
                                <NetworksFields readOnly={readOnly} />
                                <HostsFileEntriesFields readOnly={readOnly} />
                                <DNSFields readOnly={readOnly} />
                            </div>
                        </ContentBlock>
                        <ContentBlock label="Endpoint & Port Config">
                            <div className="flex flex-col gap-6">
                                <EndpointPortConfigFields readOnly={readOnly} />
                            </div>
                        </ContentBlock>
                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigNetworksFormRef>;
    defaultValues?: AppNetworkSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
