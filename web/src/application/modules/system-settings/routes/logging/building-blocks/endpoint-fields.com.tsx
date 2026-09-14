import type { ReactNode } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Controller, type FieldPath, useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { Checkbox, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import type { HivePaaSLoggingSettingsFormInput } from "../schemas";

import { FieldMessage } from "./field-message.com";

type FormInput = HivePaaSLoggingSettingsFormInput;
type EndpointField = "url" | "authMode" | "username" | "password" | "bearerToken" | "headers" | "tlsSkipVerify";

export function EndpointFields({ prefix, urlLabel, urlInfo }: Props) {
    const { control, register } = useFormContext<FormInput>();
    const path = (field: EndpointField) => `${prefix}.${field}` as FieldPath<FormInput>;
    const { field: authMode } = useController({ control, name: path("authMode") });
    const mode = typeof authMode.value === "string" ? authMode.value : "none";

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label={urlLabel}
                        content={urlInfo}
                    />
                }
            >
                <Input
                    {...register(path("url"))}
                    placeholder="https://"
                />
                <FieldMessage name={path("url")} />
            </InfoBlock>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Authentication"
                        content="How the collector proves itself here. Only the selected mode is stored: switching away clears the other one."
                    />
                }
            >
                <Tabs
                    value={mode}
                    onValueChange={authMode.onChange}
                >
                    <TabsList>
                        <TabsTrigger value="none">None</TabsTrigger>
                        <TabsTrigger value="basic">Basic auth</TabsTrigger>
                        <TabsTrigger value="bearer">Bearer token</TabsTrigger>
                    </TabsList>
                </Tabs>
            </InfoBlock>
            {mode === "basic" && (
                <>
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Username"
                                isRequired
                            />
                        }
                    >
                        <Input
                            {...register(path("username"))}
                            autoComplete="off"
                        />
                        <FieldMessage name={path("username")} />
                    </InfoBlock>
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Password"
                                content="Left masked means the stored password is kept."
                            />
                        }
                    >
                        <Controller
                            control={control}
                            name={path("password")}
                            render={({ field }) => (
                                <PasswordInput
                                    {...field}
                                    value={typeof field.value === "string" ? field.value : ""}
                                    autoComplete="new-password"
                                />
                            )}
                        />
                    </InfoBlock>
                </>
            )}
            {mode === "bearer" && (
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Bearer token"
                            content="Left masked means the stored token is kept."
                        />
                    }
                >
                    <Controller
                        control={control}
                        name={path("bearerToken")}
                        render={({ field }) => (
                            <PasswordInput
                                {...field}
                                value={typeof field.value === "string" ? field.value : ""}
                                autoComplete="new-password"
                            />
                        )}
                    />
                </InfoBlock>
            )}
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Headers"
                        content="Sent with every request to this endpoint. A tenant or routing header belongs here."
                    />
                }
            >
                <KeyValueList<FormInput>
                    name={path("headers")}
                    keyPlaceholder="name"
                    valuePlaceholder="value"
                    checkDuplicates
                    enableEditing
                />
            </InfoBlock>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Skip TLS verification"
                        content="Accept a certificate this endpoint cannot prove. Only for a host you control."
                    />
                }
            >
                <Controller
                    control={control}
                    name={path("tlsSkipVerify")}
                    render={({ field }) => (
                        <Checkbox
                            checked={field.value === true}
                            onCheckedChange={checked => {
                                field.onChange(checked === true);
                            }}
                        />
                    )}
                />
            </InfoBlock>
        </>
    );
}

type Props = {
    prefix: "ingest" | "query" | `forwards.${number}`;
    urlLabel: ReactNode;
    urlInfo?: ReactNode;
};
