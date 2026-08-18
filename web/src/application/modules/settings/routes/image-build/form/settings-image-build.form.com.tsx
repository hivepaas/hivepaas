import React, { useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { ImageBuildRepoCacheInfo, ImageBuildSettings } from "~/settings/domain";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    BuildWorkerFields,
    RepoCacheInfoFields,
    RepositorySourceFields,
    ResourceLimitFields,
} from "../building-blocks";
import {
    SettingsImageBuildFormSchema,
    type SettingsImageBuildFormSchemaInput,
    type SettingsImageBuildFormSchemaOutput,
    emptySettingsImageBuildFormDefaults,
} from "../schemas";
import type { SettingsImageBuildFormRef } from "../types";

import { mapSettingsImageBuildToFormInput } from "./settings-image-build.form-mappers";

type SchemaInput = SettingsImageBuildFormSchemaInput;
type SchemaOutput = SettingsImageBuildFormSchemaOutput;

export function SettingsImageBuildForm({
    ref,
    defaultValues,
    onSubmit,
    cacheInfo,
    cacheInfoControls,
    readOnly = false,
    workerNote,
}: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSettingsImageBuildToFormInput(defaultValues)
            : emptySettingsImageBuildFormDefaults,
        resolver: zodResolver(SettingsImageBuildFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapSettingsImageBuildToFormInput(defaultValues) : emptySettingsImageBuildFormDefaults,
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
                <ContentBlock label="Build Worker Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            {workerNote}
                            <BuildWorkerFields readOnly={readOnly} />
                        </div>
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Resource Limit Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <ResourceLimitFields />
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Repository Source Code">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <RepositorySourceFields cacheNote={cacheInfoControls.note} />
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

type Props = {
    ref?: React.Ref<SettingsImageBuildFormRef>;
    defaultValues?: ImageBuildSettings;
    cacheInfo?: ImageBuildRepoCacheInfo;
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
    workerNote?: React.ReactNode;
};
