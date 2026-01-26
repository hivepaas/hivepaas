import React, { type PropsWithChildren, useImperativeHandle, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { toast } from "sonner";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { FormHeader } from "../building-blocks";
import { BuildTimeEnvVars, RuntimeEnvVars } from "../content-blocks";
import {
    ProjectEnvVarsFormSchema,
    type ProjectEnvVarsFormSchemaInput,
    type ProjectEnvVarsFormSchemaOutput,
} from "../schemas";
import { type ProjectEnvVarsFormRef } from "../types";

const DEFAULTS: ProjectEnvVarsFormSchemaInput = {
    buildtime: [],
    runtime: [],
};

type SchemaInput = ProjectEnvVarsFormSchemaInput;
type SchemaOutput = ProjectEnvVarsFormSchemaOutput;

export function ProjectEnvVarsForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            ...DEFAULTS,
            ...defaultValues,
        },
        resolver: zodResolver(ProjectEnvVarsFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = methods.formState;

    const [search, setSearch] = useState("");
    const [isRevealed, setIsRevealed] = useState(false);
    const [viewMode, setViewMode] = useState<"merge" | "individual">("individual");

    function onValid(values: SchemaOutput) {
        if (!isDirty) {
            toast.info("No changes to save");
            return;
        }

        onSubmit({
            ...values,
        });
    }

    function onInvalid(errors: FieldErrors<SchemaInput>) {
        console.error(errors);
    }

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: SchemaInput) => {
                methods.reset({
                    ...DEFAULTS,
                    ...values,
                });
            },
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }

                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as keyof SchemaInput,
                        { message, type: "manual" },
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods],
    );

    return (
        <div className="single-project-env-vars-form">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();

                        void methods.handleSubmit(onValid, onInvalid)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <FormHeader
                        search={{ value: search, onChange: setSearch }}
                        isRevealed={isRevealed}
                        onRevealToggle={() => {
                            setIsRevealed(!isRevealed);
                        }}
                        viewMode={viewMode}
                        onViewModeChange={setViewMode}
                    />

                    <BuildTimeEnvVars
                        search={search}
                        viewMode={viewMode}
                        isRevealed={isRevealed}
                    />
                    <div className="h-px bg-border" />
                    <RuntimeEnvVars
                        search={search}
                        viewMode={viewMode}
                        isRevealed={isRevealed}
                    />

                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectEnvVarsFormRef>;
    defaultValues: Partial<ProjectEnvVarsFormSchemaInput>;
    onSubmit: (values: ProjectEnvVarsFormSchemaOutput) => void;
}>;
