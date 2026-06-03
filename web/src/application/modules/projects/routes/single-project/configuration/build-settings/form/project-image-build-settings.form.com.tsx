import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { ProjectImageBuildRepoCacheInfo, ProjectImageBuildSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { RepoCacheInfoFields, RepositorySourceFields, ResourceLimitFields } from "../building-blocks";
import {
    ProjectImageBuildSettingsFormSchema,
    type ProjectImageBuildSettingsFormSchemaInput,
    type ProjectImageBuildSettingsFormSchemaOutput,
    emptyProjectImageBuildSettingsFormDefaults,
} from "../schemas";
import type { ProjectImageBuildSettingsFormRef } from "../types";

import { mapProjectImageBuildSettingsToFormInput } from "./project-image-build-settings.form-mappers";

type SchemaInput = ProjectImageBuildSettingsFormSchemaInput;
type SchemaOutput = ProjectImageBuildSettingsFormSchemaOutput;

export function ProjectImageBuildSettingsForm({
    ref,
    defaultValues,
    onSubmit,
    cacheInfo,
    cacheInfoControls,
    readOnly = false,
    children,
}: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapProjectImageBuildSettingsToFormInput(defaultValues)
            : emptyProjectImageBuildSettingsFormDefaults,
        resolver: zodResolver(ProjectImageBuildSettingsFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues
                ? mapProjectImageBuildSettingsToFormInput(defaultValues)
                : emptyProjectImageBuildSettingsFormDefaults,
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
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods],
    );

    return (
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
                <ContentBlock label="Resource Limit">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            {children}
                            <ResourceLimitFields />
                        </div>
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Repository Sources">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            <RepositorySourceFields cacheNote={cacheInfoControls.note} />
                        </div>
                    </fieldset>
                    <div className="mt-6">
                        <RepoCacheInfoFields
                            hasQueried={cacheInfoControls.hasQueried}
                            cacheInfo={cacheInfo}
                            isQuerying={cacheInfoControls.isQuerying}
                            isClearing={cacheInfoControls.isClearing}
                            readOnly={cacheInfoControls.readOnly}
                            onQuery={cacheInfoControls.onQuery}
                            onClear={cacheInfoControls.onClear}
                        />
                    </div>
                </ContentBlock>

                {cacheInfoControls.footer}
            </form>
        </FormProvider>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectImageBuildSettingsFormRef>;
    defaultValues?: ProjectImageBuildSettings;
    cacheInfo?: ProjectImageBuildRepoCacheInfo;
    cacheInfoControls: {
        hasQueried: boolean;
        isQuerying: boolean;
        isClearing: boolean;
        readOnly: boolean;
        note: React.ReactNode;
        footer: React.ReactNode;
        onQuery: () => void;
        onClear: () => void;
    };
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
