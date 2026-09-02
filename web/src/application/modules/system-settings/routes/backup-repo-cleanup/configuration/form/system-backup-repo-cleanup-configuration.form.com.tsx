import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import type { SystemBackupRepoCleanupSettings } from "~/system-settings/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { EnabledConfigurationFields, EnabledField } from "../building-blocks";
import {
    type SystemBackupRepoCleanupConfigurationFormInput,
    type SystemBackupRepoCleanupConfigurationFormOutput,
    SystemBackupRepoCleanupConfigurationFormSchema,
} from "../schemas";
import type { SystemBackupRepoCleanupConfigurationFormRef } from "../types";

import {
    emptySystemBackupRepoCleanupConfigurationFormDefaults,
    mapSystemBackupRepoCleanupSettingsToFormInput,
} from "./system-backup-repo-cleanup-configuration.form-mappers";

type SchemaInput = SystemBackupRepoCleanupConfigurationFormInput;
type SchemaOutput = SystemBackupRepoCleanupConfigurationFormOutput;

function useSystemBackupRepoCleanupFormMethods(defaultValues?: SystemBackupRepoCleanupSettings) {
    return useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapSystemBackupRepoCleanupSettingsToFormInput(defaultValues)
            : emptySystemBackupRepoCleanupConfigurationFormDefaults,
        resolver: zodResolver(SystemBackupRepoCleanupConfigurationFormSchema),
        mode: "onSubmit",
    });
}

export function SystemBackupRepoCleanupConfigurationForm({
    ref,
    defaultValues,
    onSubmit,
    readOnly = false,
    children,
}: Props) {
    const methods = useSystemBackupRepoCleanupFormMethods(defaultValues);

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
                        if (readOnly) {
                            return;
                        }

                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
                    >
                        <EnabledField />
                        <EnabledConfigurationFields
                            nextRuns={defaultValues?.nextRuns ?? []}
                            readOnly={readOnly}
                        />
                    </fieldset>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SystemBackupRepoCleanupConfigurationFormRef>;
    defaultValues?: SystemBackupRepoCleanupSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
