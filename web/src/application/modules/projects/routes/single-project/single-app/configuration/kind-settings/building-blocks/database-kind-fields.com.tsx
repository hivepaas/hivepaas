import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { PasswordInput } from "@/components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

import { CredentialApplyInfo } from "./credential-apply-info.com";
import { KindSslCertSelect } from "./kind-ssl-cert-select.com";

interface Props {
    readOnly?: boolean;
}

const SSL_MODES = [
    { value: "disable", label: "Disable", description: "No SSL/TLS encryption" },
    { value: "prefer", label: "Prefer", description: "Try SSL/TLS, fallback to unencrypted" },
    { value: "require", label: "Require", description: "Require encrypted connection" },
    { value: "verify-ca", label: "Verify CA", description: "Require SSL and verify CA certificate" },
    { value: "verify-full", label: "Verify Full", description: "Verify CA certificate and hostname match" },
] as const;

export function DatabaseKindFields({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();

    const {
        field: dbName,
        fieldState: { invalid: isDbNameInvalid, error: dbNameError },
    } = useController({ control, name: "database.dbName" });

    const {
        field: username,
        fieldState: { invalid: isUsernameInvalid, error: usernameError },
    } = useController({ control, name: "database.username" });

    const {
        field: password,
        fieldState: { invalid: isPasswordInvalid, error: passwordError },
    } = useController({ control, name: "database.password" });

    const {
        field: rootPassword,
        fieldState: { invalid: isRootPasswordInvalid, error: rootPasswordError },
    } = useController({ control, name: "database.rootPassword" });

    const {
        field: sslMode,
        fieldState: { error: sslModeError },
    } = useController({ control, name: "database.sslMode" });

    const { field: tlsPassthrough } = useController({ control, name: "database.tlsPassthrough" });

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Database Name"
                        content="Initial database name to connect to or create."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...dbName}
                            value={dbName.value ?? ""}
                            onChange={dbName.onChange}
                            placeholder="e.g. my_database"
                            aria-invalid={isDbNameInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[dbNameError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Username"
                        content="Database user or application username."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...username}
                            value={username.value ?? ""}
                            onChange={username.onChange}
                            placeholder="e.g. postgres, root, app_user"
                            aria-invalid={isUsernameInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[usernameError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="User Password"
                        content="Password for the user. Leave empty if unconfigured, or keep existing masked value unchanged."
                    />
                }
            >
                <FieldGroup>
                    <Field className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                        <PasswordInput
                            {...password}
                            value={password.value ?? ""}
                            onChange={password.onChange}
                            aria-invalid={isPasswordInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[passwordError]} />
                        <CredentialApplyInfo kind="database" />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Root Password"
                        content="Administrative / root superuser password (e.g. for PostgreSQL postgres user or MySQL root user). Leave empty if unconfigured, or keep existing masked value unchanged."
                    />
                }
            >
                <FieldGroup>
                    <Field className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                        <PasswordInput
                            {...rootPassword}
                            value={rootPassword.value ?? ""}
                            onChange={rootPassword.onChange}
                            aria-invalid={isRootPasswordInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[rootPasswordError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="SSL Mode"
                        content="SSL/TLS connection verification requirement level."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Select
                            value={sslMode.value ?? "disable"}
                            onValueChange={sslMode.onChange}
                            disabled={readOnly}
                        >
                            <SelectTrigger className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                <SelectValue placeholder="Select SSL mode" />
                            </SelectTrigger>
                            <SelectContent>
                                {SSL_MODES.map(mode => (
                                    <SelectItem
                                        key={mode.value}
                                        value={mode.value}
                                    >
                                        <div className="flex flex-col text-left">
                                            <span className="font-medium">{mode.label}</span>
                                            <span className="text-xs text-muted-foreground">{mode.description}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <FieldError errors={[sslModeError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <KindSslCertSelect
                name="database.sslCert"
                readOnly={readOnly}
            />

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="TLS Passthrough"
                        content="When enabled, incoming TLS connection is passed directly to the database container without SSL termination at the edge proxy."
                    />
                }
            >
                <div className="flex items-center gap-2 pt-1.5">
                    <Checkbox
                        checked={Boolean(tlsPassthrough.value)}
                        onCheckedChange={val => {
                            if (readOnly) return;
                            tlsPassthrough.onChange(Boolean(val));
                        }}
                        disabled={readOnly}
                    />
                    <span className="text-sm text-muted-foreground">
                        Pass encrypted TLS traffic directly to container
                    </span>
                </div>
            </InfoBlock>
        </>
    );
}
