import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useForm } from "react-hook-form";
import { type AppDeploymentSettings } from "~/projects/domain";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

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
            useDefaultOnSuccess: false,
            success: data.notification?.success
                ? { id: data.notification.success.id, name: data.notification.success.name }
                : undefined,
            useDefaultOnFailure: false,
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
                buildTool: data.repoSource.buildTool,
                repoType: data.repoSource.repoType,
                repoUrl: data.repoSource.repoUrl,
                repoRef: data.repoSource.repoRef,
                credentials: { id: data.repoSource.credentials.id, name: data.repoSource.credentials.name },
                dockerfilePath: data.repoSource.dockerfilePath,
                imageName: data.repoSource.imageName,
                imageTags: data.repoSource.imageTags,
                pushToRegistry: { id: data.repoSource.pushToRegistry.id, name: data.repoSource.pushToRegistry.name },
            },
        };
    }

    return {
        ...base,
        activeMethod: EAppDeploymentMethod.Image,
        image: data.image,
        registryAuth: { id: data.registryAuth.id, name: data.registryAuth.name },
    };
}

export function AppConfigDeploymentSettingsForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapDefaultValues(defaultValues)
            : {
                  activeMethod: EAppDeploymentMethod.Image,
                  image: "",
                  registryAuth: undefined,
                  command: "",
                  workingDir: "",
                  preDeploymentCommand: "",
                  postDeploymentCommand: "",
                  notification: {
                      useDefaultOnSuccess: false,
                      useDefaultOnFailure: false,
                  },
              },
        resolver: zodResolver(AppConfigDeploymentSettingsFormSchema),
        mode: "onSubmit",
    });

    const activeMethod = methods.watch("activeMethod");

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
                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <div className="rounded-lg border bg-card p-6">
                        <h3 className="text-lg font-medium mb-4">Deployment Configuration</h3>
                        <div className="flex flex-col gap-6">
                            <MethodSelector />

                            {activeMethod === EAppDeploymentMethod.Image && <DockerImageFields />}
                            {activeMethod === EAppDeploymentMethod.Repo && <GitSourceFields />}
                        </div>
                    </div>

                    <div className="rounded-lg border bg-card p-6">
                        <h3 className="text-lg font-medium mb-4">Run Configuration</h3>
                        <div className="flex flex-col gap-6">
                            <RunConfigurationFields />
                        </div>
                    </div>

                    <div className="rounded-lg border bg-card p-6">
                        <h3 className="text-lg font-medium mb-4">Notification Configuration</h3>
                        <div className="flex flex-col gap-6">
                            <NotificationFields />
                        </div>
                    </div>

                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigDeploymentSettingsFormRef>;
    defaultValues?: AppDeploymentSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
