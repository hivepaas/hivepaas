import React, { type PropsWithChildren, useImperativeHandle, useRef, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectEnvVarsCommands } from "~/projects/data/commands/project-env-vars";
import { EnvVarsFormHeader, type FinalEnvValueItem, FinalEnvValuesDialog } from "~/projects/module-shared/components";

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

type EnvVarFormItem = SchemaInput["buildtime"][number];

function toEnvVarWire(envVars: EnvVarFormItem[]) {
    return envVars
        .filter(envVar => envVar.key.trim() !== "")
        .map(({ key, value, isLiteral }) => ({
            key: key.trim(),
            value,
            isLiteral,
        }));
}

export const ProjectEnvVarsForm = React.forwardRef<ProjectEnvVarsFormRef, Props>(function ProjectEnvVarsForm(
    { defaultValues, onSubmit, readOnly = false, children }: Props,
    ref: React.ForwardedRef<ProjectEnvVarsFormRef>,
) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

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

    const [finalValuesOpen, setFinalValuesOpen] = useState(false);
    const [finalValuesItems, setFinalValuesItems] = useState<FinalEnvValueItem[]>([]);
    const [finalValuesSectionTitle, setFinalValuesSectionTitle] = useState("Build Time Env Variables");

    const { mutate: computeEnvVars, isPending: isComputing } = ProjectEnvVarsCommands.useCompute({
        onSuccess: response => {
            setFinalValuesItems(response.data);
            setFinalValuesOpen(true);
        },
    });

    function handleShowFinalValues(section: "buildtime" | "runtime") {
        invariant(projectId, "projectId must be defined");

        const values = methods.getValues();

        if (section === "buildtime") {
            const buildtimeEnvVars = toEnvVarWire(values.buildtime);
            if (buildtimeEnvVars.length === 0) {
                toast.error("No buildtime env vars to compute");
                return;
            }

            setFinalValuesSectionTitle("Build Time Env Variables");
            computeEnvVars({
                projectID: projectId,
                buildtimeEnvVars,
            });
            return;
        }

        const runtimeEnvVars = toEnvVarWire(values.runtime);
        if (runtimeEnvVars.length === 0) {
            toast.error("No runtime env vars to compute");
            return;
        }

        setFinalValuesSectionTitle("Runtime Env Variables");
        computeEnvVars({
            projectID: projectId,
            runtimeEnvVars,
        });
    }

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
                            onShowFinalValues={() => {
                                handleShowFinalValues("buildtime");
                            }}
                        />
                        <div className="h-px bg-border" />
                        <EnvVarsBaseForm
                            search={search}
                            viewMode={viewMode}
                            isRevealed={isRevealed}
                            name="runtime"
                            title="Runtime Env Variables"
                            readOnly={readOnly}
                            onShowFinalValues={() => {
                                handleShowFinalValues("runtime");
                            }}
                        />

                        {children}
                    </fieldset>
                </form>
            </FormProvider>

            <FinalEnvValuesDialog
                open={finalValuesOpen || isComputing}
                onOpenChange={open => {
                    if (!open && !isComputing) {
                        setFinalValuesOpen(false);
                    }
                }}
                items={finalValuesItems}
                sectionTitle={finalValuesSectionTitle}
                isPending={isComputing}
            />
        </div>
    );
});

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectEnvVarsFormRef>;
    defaultValues: Partial<ProjectEnvVarsFormSchemaInput>;
    onSubmit: (values: ProjectEnvVarsFormSchemaOutput) => void;
    readOnly?: boolean;
}>;
