import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm } from "react-hook-form";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESSHKeyType } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditSSHKeyFormInput, CreateOrEditSSHKeyFormOutput } from "../schemas";
import { CreateOrEditSSHKeyFormSchema } from "../schemas";

const UNSPECIFIED_KEY_TYPE_VALUE = "__unspecified__";

const keyTypeOptions = [
    { value: ESSHKeyType.Ed25519, label: "Ed25519 (ed25519)" },
    { value: ESSHKeyType.ECP256, label: "ECDSA P256 (ec-p256)" },
    { value: ESSHKeyType.ECP384, label: "ECDSA P384 (ec-p384)" },
    { value: ESSHKeyType.ECP521, label: "ECDSA P521 (ec-p521)" },
    { value: ESSHKeyType.RSA2048, label: "RSA 2048 (rsa-2048)" },
    { value: ESSHKeyType.RSA3072, label: "RSA 3072 (rsa-3072)" },
    { value: ESSHKeyType.RSA4096, label: "RSA 4096 (rsa-4096)" },
    { value: ESSHKeyType.RSA8192, label: "RSA 8192 (rsa-8192)" },
    { value: "", label: "Unspecified" },
] as const;

export function CreateOrEditSSHKeyForm({
    isPending,
    isGenerating,
    onGenerate,
    onSubmit,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    onClose,
}: Props) {
    const form = useForm<CreateOrEditSSHKeyFormInput, unknown, CreateOrEditSSHKeyFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            keyType: initialValues?.keyType ?? ESSHKeyType.Ed25519,
            publicKey: initialValues?.publicKey ?? "",
            privateKey: initialValues?.privateKey ?? "",
            passphrase: initialValues?.passphrase ?? "",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditSSHKeyFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        clearErrors,
        setError,
        setValue,
        formState: { errors, isDirty },
    } = form;

    useEffect(() => {
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty, onHasChanges, readOnlyInherited]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: keyType,
        fieldState: { invalid: isKeyTypeInvalid },
    } = useController({ name: "keyType", control });
    const {
        field: publicKey,
        fieldState: { invalid: isPublicKeyInvalid },
    } = useController({ name: "publicKey", control });
    const {
        field: privateKey,
        fieldState: { invalid: isPrivateKeyInvalid },
    } = useController({ name: "privateKey", control });
    const {
        field: passphrase,
        fieldState: { invalid: isPassphraseInvalid },
    } = useController({ name: "passphrase", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditSSHKeyFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    async function handleGenerate() {
        if (readOnlyInherited) {
            return;
        }

        if (!keyType.value) {
            setError("keyType", { type: "manual", message: "Key type is required to generate an SSH key" });
            return;
        }

        clearErrors("keyType");

        const generatedKey = await onGenerate({
            keyType: keyType.value,
            passphrase: passphrase.value.trim() ? passphrase.value : undefined,
        });

        setValue("keyType", generatedKey.keyType, { shouldDirty: true, shouldValidate: true });
        setValue("publicKey", generatedKey.publicKey, { shouldDirty: true, shouldValidate: true });
        setValue("privateKey", generatedKey.privateKey, { shouldDirty: true, shouldValidate: true });
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditSSHKeyFormOutput>) {
        console.error(_errors);
    }

    return (
        <FormProvider {...form}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="flex flex-col gap-6"
            >
                {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                <fieldset
                    disabled={readOnlyInherited}
                    className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
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
                        title={<LabelWithInfo label="Key Type" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Select
                                    value={keyType.value || UNSPECIFIED_KEY_TYPE_VALUE}
                                    onValueChange={value => {
                                        const nextValue =
                                            value === UNSPECIFIED_KEY_TYPE_VALUE ? "" : (value as ESSHKeyType);
                                        keyType.onChange(nextValue);
                                    }}
                                >
                                    <SelectTrigger aria-invalid={isKeyTypeInvalid}>
                                        <SelectValue placeholder="Select key type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {keyTypeOptions.map(option => (
                                            <SelectItem
                                                key={option.value || UNSPECIFIED_KEY_TYPE_VALUE}
                                                value={option.value || UNSPECIFIED_KEY_TYPE_VALUE}
                                            >
                                                {option.label}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FieldError errors={[errors.keyType]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Public Key" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Textarea
                                    {...publicKey}
                                    className="min-h-24"
                                    aria-invalid={isPublicKeyInvalid}
                                />
                                <FieldError errors={[errors.publicKey]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Private Key"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Textarea
                                    {...privateKey}
                                    className="min-h-24"
                                    aria-invalid={isPrivateKeyInvalid}
                                />
                                <FieldError errors={[errors.privateKey]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Passphrase" />}
                    >
                        <FieldGroup>
                            <Field>
                                <PasswordInput
                                    value={passphrase.value}
                                    onChange={passphrase.onChange}
                                    aria-invalid={isPassphraseInvalid}
                                />
                                <FieldError errors={[errors.passphrase]} />
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

                    <div className={cn(dashedBorderBox, "text-center")}>
                        <span className="text-orange-500">Note:</span> You can use the Public Key on GitHub or GitLab.
                        If you intend to use it on GitHub, it is recommended to add it under the{" "}
                        <span className="text-orange-500">Deploy Keys</span> section of a specific repository instead of
                        adding it to the user&apos;s <span className="text-orange-500">SSH Keys</span> section to
                        enhance security.
                    </div>

                    {!readOnlyInherited && (
                        <Field>
                            <div className="flex justify-end gap-4">
                                <Button
                                    type="button"
                                    isLoading={isGenerating}
                                    onClick={() => {
                                        void handleGenerate();
                                    }}
                                >
                                    Generate
                                </Button>
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                >
                                    Save
                                </Button>
                            </div>
                        </Field>
                    )}
                </fieldset>
                {readOnlyInherited && (
                    <Field>
                        <div className="flex justify-end">
                            <Button
                                type="button"
                                onClick={onClose}
                            >
                                Close
                            </Button>
                        </div>
                    </Field>
                )}
            </form>
        </FormProvider>
    );
}

interface Props {
    isPending: boolean;
    isGenerating: boolean;
    onGenerate: (values: { keyType: ESSHKeyType; passphrase?: string }) => Promise<{
        keyType: ESSHKeyType;
        publicKey: string;
        privateKey: string;
    }>;
    onSubmit: (values: CreateOrEditSSHKeyFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditSSHKeyFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}
