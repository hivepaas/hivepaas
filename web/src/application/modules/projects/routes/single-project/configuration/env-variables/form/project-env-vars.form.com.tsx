import React, { type PropsWithChildren, useImperativeHandle, useRef, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { toast } from "sonner";
import { EnvVarsFormHeader } from "~/projects/module-shared/components";

import { EnvVarsBaseForm } from "@application/modules/projects/module-shared/form/env-vars/env-vars.form.com";

import { type ValidationException } from "@infrastructure/exceptions/validation";

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

export const ProjectEnvVarsForm = React.forwardRef<ProjectEnvVarsFormRef, Props>(function ProjectEnvVarsForm(
    { defaultValues, onSubmit, readOnly = false, children }: Props,
    ref: React.ForwardedRef<ProjectEnvVarsFormRef>,
) {
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
        <div className="single-project-env-vars-form">
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

                        <EnvVarsBaseForm
                            search={search}
                            viewMode={viewMode}
                            isRevealed={isRevealed}
                            name="buildtime"
                            title="Build Time Env Variables"
                            readOnly={readOnly}
                        />
                        <div className="h-px bg-border" />
                        <EnvVarsBaseForm
                            search={search}
                            viewMode={viewMode}
                            isRevealed={isRevealed}
                            name="runtime"
                            title="Runtime Env Variables"
                            readOnly={readOnly}
                        />

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
});

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectEnvVarsFormRef>;
    defaultValues: Partial<ProjectEnvVarsFormSchemaInput>;
    onSubmit: (values: ProjectEnvVarsFormSchemaOutput) => void;
    readOnly?: boolean;
}>;
