import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useController, useForm, useFormContext, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { AppFeatureSettings } from "~/projects/domain";

import { AppLink, ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import { DbAppsToCloneFields, PreviewCommandsFields } from "../building-blocks";
import {
    AppFeatureSettingsFormSchema,
    type AppFeatureSettingsFormSchemaInput,
    type AppFeatureSettingsFormSchemaOutput,
    DEFAULT_PREVIEW_CREATION_DELAY,
    FEATURE_SETTINGS_TITLE_WIDTH,
    emptyAppFeatureSettingsFormDefaults,
} from "../schemas";
import type { AppFeatureSettingsFormRef } from "../types";

type SchemaInput = AppFeatureSettingsFormSchemaInput;
type SchemaOutput = AppFeatureSettingsFormSchemaOutput;
type FeatureToggleFieldPath = Extract<
    FieldPath<SchemaInput>,
    | "loggingSettings.enabled"
    | "schedJobSettings.enabled"
    | "terminalSettings.enabled"
    | "previewSettings.enabled"
    | "performanceSettings.enabled"
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
        performanceSettings: {
            enabled: data.performanceSettings.enabled,
        },
        previewSettings: {
            enabled: data.previewSettings.enabled,
            creationDelay: data.previewSettings.creationDelay.trim() || DEFAULT_PREVIEW_CREATION_DELAY,
            appsToClone: data.previewSettings.appsToClone.map(app => ({
                id: app.id,
                name: app.name,
                ...(app.photo !== undefined ? { photo: app.photo } : {}),
            })),
            autoCloneApps: data.previewSettings.autoCloneApps,
            allowPRComments: data.previewSettings.allowPRComments,
            commands: data.previewSettings.commands.map(cmd => ({
                id: cmd.id,
                name: cmd.name,
            })),
        },
    };
}

function FeatureToggleField({ name }: { name: FeatureToggleFieldPath }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name });

    return (
        <InfoBlock
            title="Enabled"
            titleWidth={FEATURE_SETTINGS_TITLE_WIDTH}
        >
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function PreviewCreationDelayField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: "previewSettings.creationDelay" });

    return (
        <InfoBlock
            title="Preview Creation Delay"
            titleWidth={FEATURE_SETTINGS_TITLE_WIDTH}
        >
            <FieldGroup>
                <Field>
                    <Input
                        {...field}
                        placeholder={DEFAULT_PREVIEW_CREATION_DELAY}
                        className="max-w-[110px]"
                        aria-invalid={invalid}
                    />
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

function AutoCloneDbAppsField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "previewSettings.autoCloneApps" });

    return (
        <InfoBlock
            title="Auto Clone DB Apps on Preview Creation"
            titleWidth={FEATURE_SETTINGS_TITLE_WIDTH}
        >
            <Checkbox
                checked={field.value}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function AllowPRCommentsField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "previewSettings.allowPRComments" });

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Allow PR Comments"
                    content="Lets a comment on a pull request run /hivepaas deploy and /hivepaas cancel for this app. Off, a comment is answered with where to turn it on, and previews are made from the dashboard."
                />
            }
            titleWidth={FEATURE_SETTINGS_TITLE_WIDTH}
        >
            <div className="flex flex-col gap-2">
                <Checkbox
                    checked={field.value}
                    onCheckedChange={value => {
                        field.onChange(value === true);
                    }}
                />
                {field.value && (
                    <p className={cn(dashedBorderBox, "text-sm")}>
                        <span className="font-semibold text-orange-500">Warning:</span> A comment deploys the pull
                        request&apos;s code with this app&apos;s variables and secrets. On a public repository, anyone
                        who can comment can do it.
                    </p>
                )}
            </div>
        </InfoBlock>
    );
}

function AppPreviewWarningBox() {
    return (
        <div className={cn(dashedBorderBox, "space-y-2")}>
            <p>
                <span className="font-semibold text-orange-500">Warning:</span> Deploying a Preview App that executes
                database schema migrations against a shared database may break or crash the Main App.
            </p>
            <p>
                <span className="font-semibold">Solution:</span> Add the Database Apps you want to clone here so the
                Preview App operates on an isolated database instance.
            </p>
            <p>
                <span className="font-semibold text-orange-500">Note:</span> Selected Database Apps must have their
                Clone Settings pre-configured.
            </p>
        </div>
    );
}

export function AppFeatureSettingsForm({
    ref,
    projectID,
    env,
    appID,
    defaultValues,
    onSubmit,
    readOnly = false,
    hidePreviewSettings = false,
    children,
}: Props) {
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

    const previewSettingsEnabled = useWatch({ control: methods.control, name: "previewSettings.enabled" });

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
                        <div className={cn(dashedBorderBox)}>
                            <span className="font-semibold text-orange-500">Note:</span> By default, some features may
                            be locked to enhance security. If you need those features, you can unlock them here.
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

                        <ContentBlock label="Routes and Calls">
                            <div className="flex flex-col gap-3">
                                <FeatureToggleField name="performanceSettings.enabled" />
                                <p className="text-sm text-muted-foreground">
                                    Measures every request the app answers, by route, and every call it makes to a
                                    database, another app or an outside host - from inside its containers, by OBI
                                    (eBPF), on the nodes an administrator turns on in{" "}
                                    <AppLink.Modules
                                        to={ROUTE.systemSettings.logging.performance.$route}
                                        className="text-link"
                                    >
                                        System → Logging → Routes and Calls
                                    </AppLink.Modules>
                                    . They show in the app&apos;s Metrics, under Routes and Dependencies.
                                </p>
                            </div>
                        </ContentBlock>

                        {!hidePreviewSettings && (
                            <ContentBlock label="App Preview">
                                <div className="flex flex-col gap-6">
                                    <FeatureToggleField name="previewSettings.enabled" />
                                    {previewSettingsEnabled && (
                                        <>
                                            <AllowPRCommentsField />
                                            <PreviewCreationDelayField />
                                            <AppPreviewWarningBox />
                                            <DbAppsToCloneFields
                                                projectID={projectID}
                                                env={env}
                                                appID={appID}
                                                readOnly={readOnly}
                                            />
                                            <AutoCloneDbAppsField />
                                            <PreviewCommandsFields
                                                projectID={projectID}
                                                env={env}
                                                readOnly={readOnly}
                                            />
                                        </>
                                    )}
                                </div>
                            </ContentBlock>
                        )}

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppFeatureSettingsFormRef>;
    projectID: string;
    env: string;
    appID: string;
    defaultValues?: AppFeatureSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
    hidePreviewSettings?: boolean;
}>;
