import type { ReactNode } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Controller, type FieldPath, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Checkbox, Input } from "@/components/ui";

import type { HivePaaSLoggingSettingsFormInput } from "../schemas";

import { FieldMessage } from "./field-message.com";

type FormInput = HivePaaSLoggingSettingsFormInput;
type EndpointField = "url" | "username" | "password" | "bearerToken" | "tlsSkipVerify";

export function EndpointFields({ prefix, urlTitle, urlDescription, showBasicAuth = true }: Props) {
    const { control, register } = useFormContext<FormInput>();
    const path = (field: EndpointField) => `${prefix}.${field}` as FieldPath<FormInput>;

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={urlTitle}
                description={urlDescription}
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
                        title="Username"
                    >
                        <Input
                            {...register(path("username"))}
                            autoComplete="off"
                        />
                    </InfoBlock>
                    <InfoBlock
                        titleWidth={220}
                        title="Password"
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
                title="Bearer token"
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
                title="Skip TLS verification"
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
    urlTitle: ReactNode;
    urlDescription?: ReactNode;
    showBasicAuth?: boolean;
};
