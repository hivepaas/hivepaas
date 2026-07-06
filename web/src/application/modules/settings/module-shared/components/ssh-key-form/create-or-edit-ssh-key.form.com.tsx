import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm } from "react-hook-form";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESSHKeyKind, ESSHKeyType } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type { CreateOrEditSSHKeyFormInput, CreateOrEditSSHKeyFormOutput } from "./create-or-edit-ssh-key.form.schema";
import { CreateOrEditSSHKeyFormSchema } from "./create-or-edit-ssh-key.form.schema";

const UNSPECIFIED_KEY_TYPE_VALUE = "__unspecified__";

const kindOptions = Object.values(ESSHKeyKind);

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
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const form = useForm<CreateOrEditSSHKeyFormInput, unknown, CreateOrEditSSHKeyFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? ESSHKeyKind.Git,
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
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, onHasChanges, isReadOnly]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
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
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    async function handleGenerate() {
        if (isReadOnly) {
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
                className="min-h-0 flex flex-1 flex-col"
            >
                <div className="">
                    {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                    {readOnly && !readOnlyInherited && <PermissionReadonlyNotice />}
                    <fieldset
                        disabled={isReadOnly}
                        className={`flex flex-col gap-6 border-0 p-0 m-0 min-w-0 ${SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS}`}
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
                                    label="Type"
                                    isRequired
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <Select
                                        value={kind.value}
                                        onValueChange={value => {
                                            kind.onChange(value as ESSHKeyKind);
                                        }}
                                    >
                                        <SelectTrigger aria-invalid={isKindInvalid}>
                                            <SelectValue placeholder="Select type" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {kindOptions.map(option => (
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
                                        rows={4}
                                        maxRows={7}
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
                                        rows={4}
                                        maxRows={7}
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
                            <span className="text-orange-500">Note:</span> You can use the Public Key on GitHub or
                            GitLab. If you intend to use it on GitHub, it is recommended to add it under the{" "}
                            <span className="text-orange-500">Deploy Keys</span> section of a specific repository
                            instead of adding it to the user&apos;s <span className="text-orange-500">SSH Keys</span>{" "}
                            section to enhance security.
                        </div>
                    </fieldset>
                </div>
                {!isReadOnly && (
                    <div className="shrink-0 px-0 mt-6 pb-6">
                        <div className="flex justify-end gap-4">
                            <SettingsFormCancelAction
                                onCancel={onClose}
                                disabled={isPending}
                            />
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
                                className="min-w-[100px]"
                            >
                                Save
                            </Button>
                        </div>
                    </div>
                )}
                {isReadOnly && (
                    <div className="shrink-0 px-0 mt-6 pb-6">
                        <div className="flex justify-end">
                            <Button
                                type="button"
                                onClick={onClose}
                            >
                                Close
                            </Button>
                        </div>
                    </div>
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
    readOnly?: boolean;
    onClose?: () => void;
}
