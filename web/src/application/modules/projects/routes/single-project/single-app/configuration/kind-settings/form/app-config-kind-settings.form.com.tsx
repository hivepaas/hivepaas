import React, { type PropsWithChildren, useEffect, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useForm, useWatch } from "react-hook-form";
import { type AppKindSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import {
    CacheKindFields,
    CategorySelector,
    DatabaseKindFields,
    EngineFields,
    StorageKindFields,
    WebappKindInfo,
} from "../building-blocks";
import {
    AppConfigKindSettingsFormSchema,
    type AppConfigKindSettingsFormSchemaInput,
    type AppConfigKindSettingsFormSchemaOutput,
} from "../schemas";
import { type AppConfigKindSettingsFormRef } from "../types";

type SchemaInput = AppConfigKindSettingsFormSchemaInput;
type SchemaOutput = AppConfigKindSettingsFormSchemaOutput;

function mapDefaultValues(data?: AppKindSettings): SchemaInput {
    if (!data) {
        return {
            category: "webapp",
            engine: "",
            port: 8080,
            version: "",
        };
    }

    return {
        category: data.category,
        engine: data.engine,
        port: data.port || 8080,
        version: data.version,
        database: {
            dbName: data.database?.dbName ?? "",
            username: data.database?.username ?? "",
            password: data.database?.password ?? "",
            rootPassword: data.database?.rootPassword ?? "",
            sslMode: data.database?.sslMode ?? "disable",
            sslCert: data.database?.sslCert ? { id: data.database.sslCert.id, name: data.database.sslCert.name } : null,
            tlsPassthrough: data.database?.tlsPassthrough ?? false,
        },
        cache: {
            password: data.cache?.password ?? "",
            maxMemory: data.cache?.maxMemory ?? "",
            evictionRule: data.cache?.evictionRule ?? "",
            persistenceMode: data.cache?.persistenceMode ?? "",
            sslCert: data.cache?.sslCert ? { id: data.cache.sslCert.id, name: data.cache.sslCert.name } : null,
        },
        storage: {
            keyId: data.storage?.keyId ?? "",
            secret: data.storage?.secret ?? "",
            bucket: data.storage?.bucket ?? "",
            region: data.storage?.region ?? "",
        },
    };
}

interface Props extends PropsWithChildren {
    ref?: React.Ref<AppConfigKindSettingsFormRef>;
    defaultValues?: AppKindSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
    isRevealed?: boolean;
}

export function AppConfigKindSettingsForm({
    ref,
    defaultValues,
    onSubmit,
    readOnly = false,
    isRevealed = false,
    children,
}: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: mapDefaultValues(defaultValues),
        resolver: zodResolver(AppConfigKindSettingsFormSchema),
    });

    useEffect(() => {
        if (defaultValues) {
            methods.reset(mapDefaultValues(defaultValues));
        }
    }, [defaultValues, methods]);

    useImperativeHandle(
        ref,
        () => ({
            setValues: values => {
                Object.entries(values).forEach(([key, val]) => {
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    methods.setValue(key as any, val);
                });
            },
            onError: (error: ValidationException) => {
                if (error.errors.length === 0) {
                    return;
                }
                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as Parameters<typeof methods.setError>[0],
                        { message, type: "manual" },
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods],
    );

    const category = useWatch({ control: methods.control, name: "category" });

    return (
        <div className="pt-2">
            <RevealSecretsProvider value={{ isRevealed }}>
                <FormProvider {...methods}>
                    <form
                        onSubmit={event => {
                            event.preventDefault();
                            if (readOnly) return;
                            void methods.handleSubmit(onSubmit)(event);
                        }}
                        className="flex flex-col gap-6"
                    >
                        <fieldset
                            disabled={readOnly}
                            className="contents"
                        >
                            <ContentBlock label="Workload Kind & Engine">
                                <div className="flex flex-col gap-6">
                                    <CategorySelector readOnly={readOnly} />
                                    <EngineFields readOnly={readOnly} />
                                </div>
                            </ContentBlock>

                            {category === "webapp" && (
                                <ContentBlock label="Web Application Details">
                                    <WebappKindInfo />
                                </ContentBlock>
                            )}

                            {category === "database" && (
                                <ContentBlock label="Database Configuration">
                                    <div className="flex flex-col gap-6">
                                        <DatabaseKindFields readOnly={readOnly} />
                                    </div>
                                </ContentBlock>
                            )}

                            {category === "cache" && (
                                <ContentBlock label="Cache Configuration">
                                    <div className="flex flex-col gap-6">
                                        <CacheKindFields readOnly={readOnly} />
                                    </div>
                                </ContentBlock>
                            )}

                            {category === "storage" && (
                                <ContentBlock label="Object Storage Configuration">
                                    <div className="flex flex-col gap-6">
                                        <StorageKindFields readOnly={readOnly} />
                                    </div>
                                </ContentBlock>
                            )}

                            {children}
                        </fieldset>
                    </form>
                </FormProvider>
            </RevealSecretsProvider>
        </div>
    );
}
