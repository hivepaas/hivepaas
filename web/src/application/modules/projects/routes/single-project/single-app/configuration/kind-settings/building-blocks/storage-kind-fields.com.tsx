import { Field, FieldError, FieldGroup, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { PasswordInput } from "@/components/ui/input-password";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

interface Props {
    readOnly?: boolean;
}

export function StorageKindFields({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();

    const {
        field: keyId,
        fieldState: { invalid: isKeyIdInvalid, error: keyIdError },
    } = useController({ control, name: "storage.keyId" });

    const {
        field: secret,
        fieldState: { invalid: isSecretInvalid, error: secretError },
    } = useController({ control, name: "storage.secret" });

    const {
        field: bucket,
        fieldState: { invalid: isBucketInvalid, error: bucketError },
    } = useController({ control, name: "storage.bucket" });

    const {
        field: region,
        fieldState: { invalid: isRegionInvalid, error: regionError },
    } = useController({ control, name: "storage.region" });

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Access Key ID"
                        content="Storage access key identifier (e.g. minioadmin or custom user ID)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...keyId}
                            value={keyId.value ?? ""}
                            onChange={keyId.onChange}
                            placeholder="e.g. minioadmin"
                            aria-invalid={isKeyIdInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[keyIdError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Secret Access Key"
                        content="Private secret key for authentication. Leave as masked placeholder (••••••••) to keep existing secret."
                    />
                }
            >
                <FieldGroup>
                    <Field className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                        <PasswordInput
                            {...secret}
                            value={secret.value ?? ""}
                            onChange={secret.onChange}
                            placeholder="••••••••"
                            aria-invalid={isSecretInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[secretError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Default Bucket"
                        content="Initial or default bucket name for object storage."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...bucket}
                            value={bucket.value ?? ""}
                            onChange={bucket.onChange}
                            placeholder="e.g. my-bucket"
                            aria-invalid={isBucketInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[bucketError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Region"
                        content="Storage region identifier (e.g. us-east-1, local)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...region}
                            value={region.value ?? ""}
                            onChange={region.onChange}
                            placeholder="e.g. us-east-1"
                            aria-invalid={isRegionInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[regionError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>
        </>
    );
}
