import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppDockerApiSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    DockerApiEnabledField,
    DockerApiHostModeWarning,
    DockerApiLimitFields,
    DockerApiModeField,
    DockerApiPolicyFields,
} from "../building-blocks";
import {
    AppConfigDockerApiFormSchema,
    type AppConfigDockerApiFormSchemaInput,
    type AppConfigDockerApiFormSchemaOutput,
    emptyAppConfigDockerApiFormDefaults,
} from "../schemas";
import { type AppConfigDockerApiFormRef } from "../types";

import { mapAppDockerApiSettingsToFormInput } from "./app-config-docker-api.form-mappers";

type SchemaInput = AppConfigDockerApiFormSchemaInput;
type SchemaOutput = AppConfigDockerApiFormSchemaOutput;

export function AppConfigDockerApiForm({
    ref,
    defaultValues,
    canWriteCluster,
    onSubmit,
    readOnly = false,
    children,
}: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppDockerApiSettingsToFormInput(defaultValues)
            : emptyAppConfigDockerApiFormDefaults,
        resolver: zodResolver(AppConfigDockerApiFormSchema),
        mode: "onSubmit",
    });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapAppDockerApiSettingsToFormInput(defaultValues) : emptyAppConfigDockerApiFormDefaults,
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

    const enabled = useWatch({ control: methods.control, name: "enabled" });
    const mode = useWatch({ control: methods.control, name: "mode" });
    const hasHostMode = defaultValues?.enabled === true && defaultValues.mode === "host";

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
                        <ContentBlock label="Docker API">
                            <div className="flex flex-col gap-6">
                                <DockerApiIntro canWriteCluster={canWriteCluster} />
                                <DockerApiEnabledField />
                                {enabled && (
                                    <DockerApiModeField
                                        hostModeBlockedBy={defaultValues?.hostMode.blockedBy ?? "switch"}
                                        hasHostMode={hasHostMode}
                                    />
                                )}
                                {enabled && mode === "host" && <DockerApiHostModeWarning />}
                            </div>
                        </ContentBlock>

                        {enabled && mode === "proxy" && (
                            <>
                                <ContentBlock label="What Its Containers May Do">
                                    <DockerApiPolicyFields readOnly={readOnly} />
                                </ContentBlock>
                                <ContentBlock label="Limits">
                                    <DockerApiLimitFields
                                        defaults={
                                            defaultValues?.defaultLimits ?? { containers: 5, memory: "1gb", cpus: 1 }
                                        }
                                    />
                                </ContentBlock>
                            </>
                        )}

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

/**
 * What the screen is for, and the permission its changes take. Narrowing and
 * turning off are said to be open to anybody who may write the app, since the
 * person without Cluster Write is the one who needs to know that.
 */
function DockerApiIntro({ canWriteCluster }: { canWriteCluster: boolean }) {
    return (
        <div className={cn(dashedBorderBox, "space-y-2")}>
            <p>
                Some apps do their work by starting containers: a CI runner starts one per job, Autobase one per cluster
                operation. HivePaaS gives such an app the Docker API through a proxy on its node, which lets it do only
                what is set here.
            </p>
            <p>
                Turning access on, or letting the app do more, needs{" "}
                <span className="font-medium text-orange-500">Write</span> permission on the{" "}
                <span className="font-medium text-orange-500">Cluster</span> module
                {canWriteCluster
                    ? ", which you have."
                    : ", which you do not have. You can still let it do less, or turn access off."}{" "}
                The node&apos;s own socket takes the Privileged Apps switch and an administrator.
            </p>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigDockerApiFormRef>;
    defaultValues?: AppDockerApiSettings;
    canWriteCluster: boolean;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
