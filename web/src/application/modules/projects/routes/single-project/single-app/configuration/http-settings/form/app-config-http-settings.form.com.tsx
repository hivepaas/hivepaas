import React, { type PropsWithChildren, useEffect, useImperativeHandle, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppHttpSettings } from "~/projects/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { DomainConfigurableSections, DomainGeneralFields, DomainSelector, PathsSection } from "../building-blocks";
import {
    AppConfigHttpSettingsFormSchema,
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    emptyAppConfigHttpSettingsFormDefaults,
} from "../schemas";
import { type AppConfigHttpSettingsFormRef } from "../types";

import { mapAppHttpSettingsToFormInput } from "./app-config-http-settings.form-mappers";

type SchemaInput = AppConfigHttpSettingsFormSchemaInput;
type SchemaOutput = AppConfigHttpSettingsFormSchemaOutput;

function ConditionalDomainDetailSections({
    activeDomainIndex,
    setActiveDomainIndex,
}: {
    activeDomainIndex: number;
    setActiveDomainIndex: React.Dispatch<React.SetStateAction<number>>;
}) {
    const domains = useWatch<SchemaInput, "domains">({ name: "domains" });
    const hasDomains = domains.length > 0;
    const activeDomain = domains[activeDomainIndex];
    const hasRedirect = Boolean(activeDomain ? activeDomain.domainRedirect.trim() : "");

    useEffect(() => {
        const len = domains.length;
        if (len === 0) {
            setActiveDomainIndex(0);
            return;
        }
        setActiveDomainIndex(prev => (prev >= len ? len - 1 : prev));
    }, [domains.length, setActiveDomainIndex]);

    if (!hasDomains) {
        return null;
    }

    return (
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg text-red-500">
                Selected Domain: {domains[activeDomainIndex]?.domain ?? ""}
            </h3>
            <div className="flex flex-col gap-6 px-2">
                <DomainGeneralFields domainIndex={activeDomainIndex} />
            </div>
            {!hasRedirect && (
                <>
                    <div className="flex flex-col gap-6 px-2">
                        <DomainConfigurableSections domainIndex={activeDomainIndex} />
                    </div>

                    <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Path Configuration</h3>
                    <div className="flex flex-col gap-6 px-2">
                        <PathsSection domainIndex={activeDomainIndex} />
                    </div>
                </>
            )}
        </>
    );
}

export function AppConfigHttpSettingsForm({ ref, defaultValues, onSubmit, children }: Props) {
    const [activeDomainIndex, setActiveDomainIndex] = useState(0);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppHttpSettingsToFormInput(defaultValues)
            : emptyAppConfigHttpSettingsFormDefaults,
        resolver: zodResolver(AppConfigHttpSettingsFormSchema),
        mode: "onSubmit",
    });

    const { control } = methods;

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapAppHttpSettingsToFormInput(defaultValues) : emptyAppConfigHttpSettingsFormDefaults,
        );
        setActiveDomainIndex(0);
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

    const { field: exposePublicly } = useController({ control, name: "exposePublicly" });

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
                    <div className="flex flex-col gap-6 px-2">
                        <DomainSelector
                            activeDomainIndex={activeDomainIndex}
                            setActiveDomainIndex={setActiveDomainIndex}
                            internalEndpoints={defaultValues?.internalEndpoints ?? []}
                            domainSuggestion={defaultValues?.domainSuggestion ?? ""}
                        />
                    </div>

                    {exposePublicly.value && (
                        <ConditionalDomainDetailSections
                            activeDomainIndex={activeDomainIndex}
                            setActiveDomainIndex={setActiveDomainIndex}
                        />
                    )}

                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigHttpSettingsFormRef>;
    defaultValues?: AppHttpSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
