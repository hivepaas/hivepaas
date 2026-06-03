import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { ProjectDomainSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import { AllowedDomainsFields, CertificateConfigurationFields } from "../building-blocks";
import {
    ProjectDomainSettingsFormSchema,
    type ProjectDomainSettingsFormSchemaInput,
    type ProjectDomainSettingsFormSchemaOutput,
    emptyProjectDomainSettingsFormDefaults,
} from "../schemas";
import type { ProjectDomainSettingsFormRef } from "../types";

import { mapProjectDomainSettingsToFormInput } from "./project-domain-settings.form-mappers";

type SchemaInput = ProjectDomainSettingsFormSchemaInput;
type SchemaOutput = ProjectDomainSettingsFormSchemaOutput;

export function ProjectDomainSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, note, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapProjectDomainSettingsToFormInput(defaultValues)
            : emptyProjectDomainSettingsFormDefaults,
        resolver: zodResolver(ProjectDomainSettingsFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapProjectDomainSettingsToFormInput(defaultValues) : emptyProjectDomainSettingsFormDefaults,
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
                <ContentBlock label="Domain Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            {note}
                            <AllowedDomainsFields readOnly={readOnly} />
                        </div>
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Certificate Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <CertificateConfigurationFields readOnly={readOnly} />
                    </fieldset>
                </ContentBlock>

                {children}
            </form>
        </FormProvider>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectDomainSettingsFormRef>;
    defaultValues?: ProjectDomainSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
    note?: React.ReactNode;
}>;
