import React, { type PropsWithChildren, useImperativeHandle, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { toast } from "sonner";
import { EnvVarsFormHeader } from "~/projects/module-shared/components";
import { EnvVarsBaseForm, InheritedEnvVarsAccordion } from "~/projects/module-shared/form";
import { type EnvVarsFormBaseSchemaInput } from "~/projects/module-shared/schemas";

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
    { defaultValues, inheritedValues, onSubmit, readOnly = false, children }: Props,
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
    const [sortOrder, setSortOrder] = useState<"normal" | "asc" | "desc">("normal");
    const originalOrderRef = useRef<{ buildtime: SchemaInput["buildtime"]; runtime: SchemaInput["runtime"] } | null>(
        null,
    );

    function handleSortCycle() {
        const next = sortOrder === "normal" ? "asc" : sortOrder === "asc" ? "desc" : "normal";

        if (next === "asc" || next === "desc") {
            originalOrderRef.current ??= {
                buildtime: methods.getValues("buildtime"),
                runtime: methods.getValues("runtime"),
            };
            const sorter = (a: { key: string }, b: { key: string }) =>
                next === "asc" ? a.key.localeCompare(b.key) : b.key.localeCompare(a.key);
            methods.setValue("buildtime", [...methods.getValues("buildtime")].sort(sorter), { shouldDirty: true });
            methods.setValue("runtime", [...methods.getValues("runtime")].sort(sorter), { shouldDirty: true });
        } else {
            if (originalOrderRef.current) {
                methods.setValue("buildtime", originalOrderRef.current.buildtime, { shouldDirty: true });
                methods.setValue("runtime", originalOrderRef.current.runtime, { shouldDirty: true });
                originalOrderRef.current = null;
            }
        }
        setSortOrder(next);
    }

    function onValid(values: SchemaOutput) {
        if (readOnly) {
            return;
        }

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
                        if (readOnly) {
                            return;
                        }

                        void methods.handleSubmit(onValid, onInvalid)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                            <p>
                                <span className="text-orange-500">Note:</span> From an env var, you can reference
                                another env var or secret, for example:{" "}
                                <span className="text-orange-500">MY_ENV = {"${ENV2}"}</span> or{" "}
                                <span className="text-orange-500">MY_ENV = {"${secrets.MY_SECRET}"}</span>.
                            </p>
                            <p className="mt-3">
                                Use Secrets if you do not want anyone to see their values. Secrets will be filtered out
                                from log data, while Env vars will not.
                            </p>
                        </div>

                        <EnvVarsFormHeader
                            search={{ value: search, onChange: setSearch }}
                            isRevealed={isRevealed}
                            onRevealToggle={() => {
                                setIsRevealed(!isRevealed);
                            }}
                            viewMode={viewMode}
                            onViewModeChange={setViewMode}
                            onSortCycle={handleSortCycle}
                            sortOrder={sortOrder}
                            readOnly={readOnly}
                        />

                        {inheritedValues && (
                            <InheritedEnvVarsAccordion
                                title="Inherited Build Time Env Variables"
                                items={inheritedValues.buildtime}
                                search={search}
                                isRevealed={isRevealed}
                            />
                        )}
                        <EnvVarsBaseForm
                            search={search}
                            viewMode={viewMode}
                            isRevealed={isRevealed}
                            name="buildtime"
                            title="Buildtime Env Vars"
                            readOnly={readOnly}
                        />
                        <div className="h-px bg-border" />
                        {inheritedValues && (
                            <InheritedEnvVarsAccordion
                                title="Inherited Runtime Env Variables"
                                items={inheritedValues.runtime}
                                search={search}
                                isRevealed={isRevealed}
                            />
                        )}
                        <EnvVarsBaseForm
                            search={search}
                            viewMode={viewMode}
                            isRevealed={isRevealed}
                            name="runtime"
                            title="Runtime Env Vars"
                            readOnly={readOnly}
                        />

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
});

type EnvVarRecord = EnvVarsFormBaseSchemaInput["buildtime"][number];

type Props = PropsWithChildren<{
    defaultValues: Partial<AppConfigEnvVarsFormSchemaInput>;
    inheritedValues?: { buildtime: EnvVarRecord[]; runtime: EnvVarRecord[] };
    onSubmit: (values: AppConfigEnvVarsFormSchemaOutput) => void;
    readOnly?: boolean;
}>;
