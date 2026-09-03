import { useMemo } from "react";

import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESslCertType, ESslKeyType } from "@application/shared/enums";

import {
    Field,
    FieldError,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";

import { type ProjectDomainSettingsFormSchemaInput, ProjectDomainSettingsKeyTypeUnspecified } from "../schemas";

const SSL_CERT_TYPES: { value: ESslCertType; label: string }[] = [
    { value: ESslCertType.LetsEncrypt, label: "Let's Encrypt" },
    { value: ESslCertType.GoogleTrust, label: "Google Trust" },
    { value: ESslCertType.ZeroSSL, label: "Zero SSL" },
    { value: ESslCertType.SelfSigned, label: "Self-Signed" },
    { value: ESslCertType.Custom, label: "Custom" },
];

const SSL_KEY_TYPES: ESslKeyType[] = [
    ESslKeyType.ECP256,
    ESslKeyType.ECP384,
    ESslKeyType.ECP521,
    ESslKeyType.RSA2048,
    ESslKeyType.RSA3072,
    ESslKeyType.RSA4096,
];

function formatKeyTypeLabel(value: ESslKeyType): string {
    switch (value) {
        case ESslKeyType.ECP256:
            return "ECDSA P256 (ec-p256)";
        case ESslKeyType.ECP384:
            return "ECDSA P384 (ec-p384)";
        case ESslKeyType.ECP521:
            return "ECDSA P521 (ec-p521)";
        case ESslKeyType.RSA2048:
            return "RSA 2048 (rsa-2048)";
        case ESslKeyType.RSA3072:
            return "RSA 3072 (rsa-3072)";
        case ESslKeyType.RSA4096:
            return "RSA 4096 (rsa-4096)";
        default:
            return value;
    }
}

export function CertificateConfigurationFields({ readOnly = false }: Props) {
    const {
        control,
        formState: { errors },
    } = useFormContext<ProjectDomainSettingsFormSchemaInput>();
    const {
        field: certType,
        fieldState: { invalid: isCertTypeInvalid },
    } = useController({ name: "certSettings.certType", control });
    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({
        name: "certSettings.email",
        control,
    });
    const {
        field: keyType,
        fieldState: { invalid: isKeyTypeInvalid },
    } = useController({
        name: "certSettings.keyType",
        control,
    });

    const keyTypeOptions = useMemo(
        () =>
            SSL_KEY_TYPES.map(value => ({
                value,
                label: formatKeyTypeLabel(value),
            })),
        [],
    );

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Default Cert Type"
                        content="Default certificate type for project domains."
                    />
                }
            >
                <Field>
                    <Select
                        value={certType.value}
                        onValueChange={value => {
                            certType.onChange(value as ESslCertType);
                        }}
                        disabled={readOnly}
                    >
                        <SelectTrigger
                            aria-invalid={isCertTypeInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        >
                            <SelectValue placeholder="Select cert type" />
                        </SelectTrigger>
                        <SelectContent>
                            {SSL_CERT_TYPES.map(option => (
                                <SelectItem
                                    key={option.value}
                                    value={option.value}
                                >
                                    {option.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <FieldError errors={[errors.certSettings?.certType]} />
                </Field>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Default Registration Email"
                        content="Default registration email for Let's Encrypt certificates."
                    />
                }
            >
                <Field>
                    <Input
                        {...email}
                        type="email"
                        placeholder="email address"
                        aria-invalid={isEmailInvalid}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        disabled={readOnly}
                    />
                    <FieldError errors={[errors.certSettings?.email]} />
                </Field>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Default Key Type"
                        content="Default private key type for generated certificates."
                    />
                }
            >
                <Field>
                    <Select
                        value={keyType.value}
                        onValueChange={value => {
                            keyType.onChange(value);
                        }}
                        disabled={readOnly}
                    >
                        <SelectTrigger
                            aria-invalid={isKeyTypeInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        >
                            <SelectValue placeholder="Unspecified" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={ProjectDomainSettingsKeyTypeUnspecified}>Unspecified</SelectItem>
                            {keyTypeOptions.map(option => (
                                <SelectItem
                                    key={option.value}
                                    value={option.value}
                                >
                                    {option.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <FieldError errors={[errors.certSettings?.keyType]} />
                </Field>
            </InfoBlock>
        </div>
    );
}

type Props = {
    readOnly?: boolean;
};
