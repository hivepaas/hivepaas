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
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";

import { type ProjectDomainSettingsFormSchemaInput, ProjectDomainSettingsKeyTypeUnspecified } from "../schemas";

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
    const { field: certType } = useController({ name: "certSettings.certType", control });
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
                title={
                    <LabelWithInfo
                        label="Default Cert Type"
                        content="Default certificate type for project domains."
                    />
                }
            >
                <Tabs
                    value={certType.value}
                    onValueChange={value => {
                        certType.onChange(value);
                    }}
                >
                    <TabsList>
                        <TabsTrigger
                            value={ESslCertType.LetsEncrypt}
                            disabled={readOnly}
                        >
                            Let&apos;s Encrypt
                        </TabsTrigger>
                        <TabsTrigger
                            value={ESslCertType.Custom}
                            disabled={readOnly}
                        >
                            Custom
                        </TabsTrigger>
                    </TabsList>
                </Tabs>
            </InfoBlock>

            <InfoBlock
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
