import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ECloudStorageKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditCloudStorageFormInput, CreateOrEditCloudStorageFormOutput } from "../schemas";
import { CreateOrEditCloudStorageFormSchema } from "../schemas";

const providerOptions = Object.values(ECloudStorageKind);

export function CreateOrEditCloudStorageForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestConnection,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditCloudStorageFormInput, unknown, CreateOrEditCloudStorageFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? ECloudStorageKind.AWSS3,
            accessKeyId: initialValues?.accessKeyId ?? "",
            secretKey: initialValues?.secretKey ?? "",
            region: initialValues?.region ?? "",
            bucket: initialValues?.bucket ?? "",
            endpoint: initialValues?.endpoint ?? "",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditCloudStorageFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: accessKeyId,
        fieldState: { invalid: isAccessKeyIdInvalid },
    } = useController({ name: "accessKeyId", control });
    const {
        field: secretKey,
        fieldState: { invalid: isSecretKeyInvalid },
    } = useController({ name: "secretKey", control });
    const {
        field: region,
        fieldState: { invalid: isRegionInvalid },
    } = useController({ name: "region", control });
    const {
        field: bucket,
        fieldState: { invalid: isBucketInvalid },
    } = useController({ name: "bucket", control });
    const {
        field: endpoint,
        fieldState: { invalid: isEndpointInvalid },
    } = useController({ name: "endpoint", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditCloudStorageFormOutput) {
        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditCloudStorageFormOutput) {
        onTestConnection(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditCloudStorageFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="flex flex-col gap-6"
        >
            <InfoBlock
                titleWidth={220}
                title={<LabelWithInfo label="Name" />}
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...name}
                            aria-invalid={isNameInvalid}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Provider"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Select
                            value={kind.value}
                            onValueChange={kind.onChange}
                        >
                            <SelectTrigger aria-invalid={isKindInvalid}>
                                <SelectValue placeholder="Select provider" />
                            </SelectTrigger>
                            <SelectContent>
                                {providerOptions.map(option => (
                                    <SelectItem
                                        key={option}
                                        value={option}
                                    >
                                        {option}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <FieldError errors={[errors.kind]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Access Key ID"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...accessKeyId}
                            aria-invalid={isAccessKeyIdInvalid}
                        />
                        <FieldError errors={[errors.accessKeyId]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Secret Key"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <PasswordInput
                            value={secretKey.value}
                            onChange={secretKey.onChange}
                            aria-invalid={isSecretKeyInvalid}
                        />
                        <FieldError errors={[errors.secretKey]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Region"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...region}
                            aria-invalid={isRegionInvalid}
                        />
                        <FieldError errors={[errors.region]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Bucket"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...bucket}
                            aria-invalid={isBucketInvalid}
                        />
                        <FieldError errors={[errors.bucket]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={<LabelWithInfo label="Endpoint" />}
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...endpoint}
                            aria-invalid={isEndpointInvalid}
                        />
                        <FieldError errors={[errors.endpoint]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            {showAvailableInProjects && (
                <InfoBlock
                    titleWidth={220}
                    title={<LabelWithInfo label="Available in Projects" />}
                >
                    <Checkbox
                        checked={availableInProjects.value}
                        onCheckedChange={checked => {
                            availableInProjects.onChange(Boolean(checked));
                        }}
                    />
                </InfoBlock>
            )}

            <InfoBlock
                titleWidth={220}
                title={<LabelWithInfo label="Default" />}
            >
                <Checkbox
                    checked={defaultField.value}
                    onCheckedChange={checked => {
                        defaultField.onChange(Boolean(checked));
                    }}
                />
            </InfoBlock>

            <Field>
                <div className="flex items-center justify-between gap-4">
                    <div className="flex items-center gap-3">
                        <Button
                            type="button"
                            variant="secondary"
                            isLoading={isTesting}
                            onClick={() => {
                                void handleSubmit(onTestValid, onInvalid)();
                            }}
                        >
                            Test Access
                        </Button>
                        {testStatus === "succeeded" && <span className="text-sm text-green-600">Succeeded</span>}
                        {testStatus === "failed" && <span className="text-sm text-destructive">Failed</span>}
                    </div>
                    <Button
                        type="submit"
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </Field>
        </form>
    );
}

interface Props {
    isPending: boolean;
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditCloudStorageFormOutput) => void;
    onTestConnection: (values: CreateOrEditCloudStorageFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditCloudStorageFormInput>;
    showAvailableInProjects: boolean;
}
