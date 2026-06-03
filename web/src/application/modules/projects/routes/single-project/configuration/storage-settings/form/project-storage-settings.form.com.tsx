import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { ProjectStorageSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import {
    AllowedBaseDirectoriesFields,
    AllowedVolumesFields,
    EnabledField,
    RequiredSubpathField,
    TmpfsMaxSizeField,
} from "../building-blocks";
import {
    ProjectStorageSettingsFormSchema,
    type ProjectStorageSettingsFormSchemaInput,
    type ProjectStorageSettingsFormSchemaOutput,
    emptyProjectStorageSettingsFormDefaults,
} from "../schemas";
import type { ProjectStorageSettingsFormRef } from "../types";

import { mapProjectStorageSettingsToFormInput } from "./project-storage-settings.form-mappers";

type SchemaInput = ProjectStorageSettingsFormSchemaInput;
type SchemaOutput = ProjectStorageSettingsFormSchemaOutput;

export function ProjectStorageSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, note, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapProjectStorageSettingsToFormInput(defaultValues)
            : emptyProjectStorageSettingsFormDefaults,
        resolver: zodResolver(ProjectStorageSettingsFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues
                ? mapProjectStorageSettingsToFormInput(defaultValues)
                : emptyProjectStorageSettingsFormDefaults,
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

    const bindSettingsEnabled = useWatch({ control: methods.control, name: "bindSettings.enabled" });
    const volumeSettingsEnabled = useWatch({ control: methods.control, name: "volumeSettings.enabled" });
    const clusterVolumeSettingsEnabled = useWatch({
        control: methods.control,
        name: "clusterVolumeSettings.enabled",
    });
    const tmpfsSettingsEnabled = useWatch({ control: methods.control, name: "tmpfsSettings.enabled" });

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
                {note}

                <ContentBlock label="Bind Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            <EnabledField name="bindSettings.enabled" />
                            {bindSettingsEnabled && (
                                <>
                                    <AllowedBaseDirectoriesFields readOnly={readOnly} />
                                    <RequiredSubpathField name="bindSettings.subpathTemplate" />
                                </>
                            )}
                        </div>
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Volume Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            <EnabledField name="volumeSettings.enabled" />
                            {volumeSettingsEnabled && (
                                <>
                                    <AllowedVolumesFields
                                        name="volumeSettings.volumes"
                                        type="volume"
                                        readOnly={readOnly}
                                    />
                                    <RequiredSubpathField name="volumeSettings.subpathTemplate" />
                                </>
                            )}
                        </div>
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Cluster Volume Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            <EnabledField name="clusterVolumeSettings.enabled" />
                            {clusterVolumeSettingsEnabled && (
                                <>
                                    <AllowedVolumesFields
                                        name="clusterVolumeSettings.volumes"
                                        type="cluster"
                                        readOnly={readOnly}
                                    />
                                    <RequiredSubpathField name="clusterVolumeSettings.subpathTemplate" />
                                </>
                            )}
                        </div>
                    </fieldset>
                </ContentBlock>

                <ContentBlock label="Tmpfs Configuration">
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6">
                            <EnabledField name="tmpfsSettings.enabled" />
                            {tmpfsSettingsEnabled && <TmpfsMaxSizeField />}
                        </div>
                    </fieldset>
                </ContentBlock>

                {children}
            </form>
        </FormProvider>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectStorageSettingsFormRef>;
    defaultValues?: ProjectStorageSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
    note?: React.ReactNode;
}>;
