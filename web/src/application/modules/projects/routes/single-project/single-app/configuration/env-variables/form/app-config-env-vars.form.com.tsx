import React, { type PropsWithChildren, useImperativeHandle, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { toast } from "sonner";
import { EnvVarsFormHeader } from "~/projects/module-shared/components";
import { EnvVarsBaseForm } from "~/projects/module-shared/form";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    AppConfigEnvVarsFormSchema,
    type AppConfigEnvVarsFormSchemaInput,
    type AppConfigEnvVarsFormSchemaOutput,
} from "../schemas";
import { type AppConfigEnvVarsFormRef } from "../types";

const DEFAULTS: AppConfigEnvVarsFormSchemaInput = {
    buildtime: [],
    runtime: [],
};

type SchemaInput = AppConfigEnvVarsFormSchemaInput;
type SchemaOutput = AppConfigEnvVarsFormSchemaOutput;

export const AppConfigEnvVarsForm = React.forwardRef<AppConfigEnvVarsFormRef, Props>(function AppConfigEnvVarsForm(
    { defaultValues, onSubmit, children }: Props,
    ref: React.ForwardedRef<AppConfigEnvVarsFormRef>,
) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            ...DEFAULTS,
            ...defaultValues,
        },
        resolver: zodResolver(AppConfigEnvVarsFormSchema),
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
        <div className="single-app-env-vars-form">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();

                        void methods.handleSubmit(onValid, onInvalid)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <EnvVarsFormHeader
                        search={{ value: search, onChange: setSearch }}
                        isRevealed={isRevealed}
                        onRevealToggle={() => {
                            setIsRevealed(!isRevealed);
                        }}
                        viewMode={viewMode}
                        onViewModeChange={setViewMode}
                    />

                    <EnvVarsBaseForm
                        search={search}
                        viewMode={viewMode}
                        isRevealed={isRevealed}
                        name="buildtime"
                        title="Buildtime Env Vars"
                    />
                    <div className="h-px bg-border" />
                    <EnvVarsBaseForm
                        search={search}
                        viewMode={viewMode}
                        isRevealed={isRevealed}
                        name="runtime"
                        title="Runtime Env Vars"
                    />

                    {children}
                </form>
            </FormProvider>
        </div>
    );
});

type Props = PropsWithChildren<{
    defaultValues: Partial<AppConfigEnvVarsFormSchemaInput>;
    onSubmit: (values: AppConfigEnvVarsFormSchemaOutput) => void;
}>;
