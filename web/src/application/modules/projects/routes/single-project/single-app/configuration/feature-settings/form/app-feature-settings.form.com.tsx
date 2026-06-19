import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useController, useForm, useFormContext } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppFeatureSettings } from "~/projects/domain";

import { ContentBlock, InfoBlock } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import {
    AppFeatureSettingsFormSchema,
    type AppFeatureSettingsFormSchemaInput,
    type AppFeatureSettingsFormSchemaOutput,
    emptyAppFeatureSettingsFormDefaults,
} from "../schemas";
import type { AppFeatureSettingsFormRef } from "../types";

type SchemaInput = AppFeatureSettingsFormSchemaInput;
type SchemaOutput = AppFeatureSettingsFormSchemaOutput;
type FeatureToggleFieldPath = Extract<
    FieldPath<SchemaInput>,
    "loggingSettings.enabled" | "schedJobSettings.enabled" | "terminalSettings.enabled"
>;

function mapFeatureSettingsToFormInput(data: AppFeatureSettings): SchemaInput {
    return {
        loggingSettings: {
            enabled: data.loggingSettings.enabled,
        },
        schedJobSettings: {
            enabled: data.schedJobSettings.enabled,
        },
        terminalSettings: {
            enabled: data.terminalSettings.enabled,
        },
    };
}

function FeatureToggleField({ name }: { name: FeatureToggleFieldPath }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name });

    return (
        <InfoBlock title="Enabled">
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

export function AppFeatureSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapFeatureSettingsToFormInput(defaultValues)
            : emptyAppFeatureSettingsFormDefaults,
        resolver: zodResolver(AppFeatureSettingsFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapFeatureSettingsToFormInput(defaultValues) : emptyAppFeatureSettingsFormDefaults,
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
                        className="contents"
                    >
                        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                            <span className="text-orange-500">Note:</span> By default, some features may be locked to
                            enhance security. If you need those features, you can unlock them here.
                        </div>

                        <ContentBlock label="Service logs">
                            <FeatureToggleField name="loggingSettings.enabled" />
                        </ContentBlock>

                        <ContentBlock label="Scheduled Jobs">
                            <FeatureToggleField name="schedJobSettings.enabled" />
                        </ContentBlock>

                        <ContentBlock label="Terminal">
                            <FeatureToggleField name="terminalSettings.enabled" />
                        </ContentBlock>

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppFeatureSettingsFormRef>;
    defaultValues?: AppFeatureSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
