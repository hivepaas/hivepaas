import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { type AppDeploymentSettings } from "~/projects/domain";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    DockerImageFields,
    GitSourceFields,
    MethodSelector,
    NotificationFields,
    RunConfigurationFields,
} from "../building-blocks";
import {
    AppConfigDeploymentSettingsFormSchema,
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";
import { type AppConfigDeploymentSettingsFormRef } from "../types";

type SchemaInput = AppConfigDeploymentSettingsFormSchemaInput;
type SchemaOutput = AppConfigDeploymentSettingsFormSchemaOutput;

function mapDefaultValues(data: AppDeploymentSettings): SchemaInput {
    const base = {
        command: data.command ?? "",
        workingDir: data.workingDir ?? "",
        preDeploymentCommand: data.preDeploymentCommand ?? "",
        postDeploymentCommand: data.postDeploymentCommand ?? "",
        notification: {
            successUseDefault: data.notification?.successUseDefault ?? true,
            success: data.notification?.success
                ? { id: data.notification.success.id, name: data.notification.success.name }
                : undefined,
            failureUseDefault: data.notification?.failureUseDefault ?? true,
            failure: data.notification?.failure
                ? { id: data.notification.failure.id, name: data.notification.failure.name }
                : undefined,
        },
    };

    if (data.activeMethod === EAppDeploymentMethod.Repo) {
        return {
            ...base,
            activeMethod: EAppDeploymentMethod.Repo,
            repoSource: {
                // buildTool: data.repoSource.buildTool ?? EBuildTool.Docker,
                // repoType: data.repoSource.repoType ?? ERepoType.Git,
                repoUrl: data.repoSource.repoUrl,
                repoRef: data.repoSource.repoRef,
                commitHash: data.repoSource.commitHash,
                repoOptions: {
                    gitSubmodulesEnabled: data.repoSource.repoOptions.gitSubmodulesEnabled,
                    gitLfsEnabled: data.repoSource.repoOptions.gitLfsEnabled,
                },
                credentials: data.repoSource.credentials
                    ? {
                          id: data.repoSource.credentials.id,
                          name: data.repoSource.credentials.name,
                          type: data.repoSource.credentials.type,
                      }
                    : undefined,
                dockerfilePath: data.repoSource.dockerfilePath,
                imageName: data.repoSource.imageName,
                imageTags: data.repoSource.imageTags,
                pushToRegistry: data.repoSource.pushToRegistry
                    ? { id: data.repoSource.pushToRegistry.id, name: data.repoSource.pushToRegistry.name }
                    : undefined,
            },
        };
    }

    return {
        ...base,
        activeMethod: EAppDeploymentMethod.Image,
        imageSource: {
            image: data.imageSource.image,
            registryAuth: data.imageSource.registryAuth
                ? { id: data.imageSource.registryAuth.id, name: data.imageSource.registryAuth.name }
                : undefined,
        },
    };
}

export function AppConfigDeploymentSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapDefaultValues(defaultValues)
            : {
                  activeMethod: EAppDeploymentMethod.Image,
                  imageSource: { image: "", registryAuth: undefined },
                  command: "",
                  workingDir: "",
                  preDeploymentCommand: "",
                  postDeploymentCommand: "",
                  notification: {
                      successUseDefault: true,
                      failureUseDefault: true,
                  },
              },
        resolver: zodResolver(AppConfigDeploymentSettingsFormSchema),
        mode: "onSubmit",
    });

    const activeMethod = methods.watch("activeMethod");

    function onValid(values: SchemaOutput) {
        if (readOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(errors: FieldErrors<SchemaInput>) {
        console.error(errors);
    }

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
        <div className="pt-2">
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
                        <ContentBlock label="Deployment Configuration">
                            <div className="flex flex-col gap-6">
                                <MethodSelector readOnly={readOnly} />

                                {activeMethod === EAppDeploymentMethod.Image && (
                                    <DockerImageFields readOnly={readOnly} />
                                )}
                                {activeMethod === EAppDeploymentMethod.Repo && <GitSourceFields readOnly={readOnly} />}
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Run Configuration">
                            <div className="flex flex-col gap-6">
                                <RunConfigurationFields />
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Notification Configuration">
                            <div className="flex flex-col gap-6">
                                <NotificationFields readOnly={readOnly} />
                            </div>
                        </ContentBlock>

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigDeploymentSettingsFormRef>;
    defaultValues?: AppDeploymentSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
