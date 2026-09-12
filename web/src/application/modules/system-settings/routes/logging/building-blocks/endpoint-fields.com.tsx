import type { ReactNode } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Controller, type FieldPath, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Checkbox, Input } from "@/components/ui";

import type { HivePaaSLoggingSettingsFormInput } from "../schemas";

import { FieldMessage } from "./field-message.com";

type FormInput = HivePaaSLoggingSettingsFormInput;
type EndpointField = "url" | "username" | "password" | "bearerToken" | "tlsSkipVerify";

export function EndpointFields({ prefix, urlLabel, urlInfo, showBasicAuth = true }: Props) {
    const { control, register } = useFormContext<FormInput>();
    const path = (field: EndpointField) => `${prefix}.${field}` as FieldPath<FormInput>;

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
            {showBasicAuth && (
                <>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Username" />}
                    >
                        <Input
                            {...register(path("username"))}
                            autoComplete="off"
                        />
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
    showBasicAuth?: boolean;
};
