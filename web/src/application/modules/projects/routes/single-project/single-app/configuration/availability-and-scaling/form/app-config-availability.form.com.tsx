import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppServiceSettings, ServiceModeSpec } from "~/projects/domain";

import { EServiceMode } from "@application/modules/projects/module-shared/enums";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { PlacementConstraintsFields, PlacementPreferencesFields, ServiceModeFields } from "../building-blocks";
import {
    AppConfigAvailabilitySchema,
    type AppConfigAvailabilitySchemaInput,
    type AppConfigAvailabilitySchemaOutput,
} from "../schemas";
import { type AppConfigAvailabilityFormRef } from "../types";

type SchemaInput = AppConfigAvailabilitySchemaInput;
type SchemaOutput = AppConfigAvailabilitySchemaOutput;

export function AppConfigAvailabilityForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            mode: defaultValues?.modeSpec.mode ?? EServiceMode.Global,
            serviceReplicas:
                defaultValues?.modeSpec.mode === EServiceMode.Replicated
                    ? defaultValues.modeSpec.serviceReplicas
                    : null,
            jobMaxConcurrent:
                defaultValues?.modeSpec.mode === EServiceMode.ReplicatedJob
                    ? defaultValues.modeSpec.jobMaxConcurrent
                    : null,
            jobTotalCompletions:
                defaultValues?.modeSpec.mode === EServiceMode.ReplicatedJob
                    ? defaultValues.modeSpec.jobTotalCompletions
                    : null,
            constraints: defaultValues?.placement?.constraints ?? [],
            preferences: defaultValues?.placement?.preferences ?? [],
        },
        resolver: zodResolver(AppConfigAvailabilitySchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset({
            mode: defaultValues?.modeSpec.mode ?? EServiceMode.Global,
            serviceReplicas:
                defaultValues?.modeSpec.mode === EServiceMode.Replicated
                    ? defaultValues.modeSpec.serviceReplicas
                    : null,
            jobMaxConcurrent:
                defaultValues?.modeSpec.mode === EServiceMode.ReplicatedJob
                    ? defaultValues.modeSpec.jobMaxConcurrent
                    : null,
            jobTotalCompletions:
                defaultValues?.modeSpec.mode === EServiceMode.ReplicatedJob
                    ? defaultValues.modeSpec.jobTotalCompletions
                    : null,
            constraints: defaultValues?.placement?.constraints ?? [],
            preferences: defaultValues?.placement?.preferences ?? [],
        });
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
                    className="flex flex-col gap-3"
                >
                    <ServiceModeFields />
                    <div className="h-px bg-zinc-200" />
                    <PlacementConstraintsFields />
                    <div className="h-px bg-zinc-200" />
                    <PlacementPreferencesFields />
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigAvailabilityFormRef>;
    defaultValues?: AppServiceSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
