import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppResourceSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    CapabilitiesFields,
    MemoryFields,
    ResourceLimitFields,
    ResourceReservationFields,
} from "../building-blocks";
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

export function AppConfigResourcesForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
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
                        <ContentBlock label="Resource Reservation">
                            <ResourceReservationFields />
                        </ContentBlock>

                        <ContentBlock label="Resource Limit">
                            <ResourceLimitFields />
                        </ContentBlock>
                        <ContentBlock label="Memory">
                            <MemoryFields />
                        </ContentBlock>
                        <ContentBlock label="Capabilities">
                            <CapabilitiesFields />
                        </ContentBlock>

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigResourcesFormRef>;
    defaultValues?: AppResourceSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
