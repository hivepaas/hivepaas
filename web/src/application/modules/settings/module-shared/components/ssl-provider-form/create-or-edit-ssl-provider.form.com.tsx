import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { SSL_KEY_TYPE_OPTIONS, SSL_PROVIDER_OPTIONS } from "~/settings/module-shared/constants/ssl-provider.constants";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESslProviderKind } from "@application/shared/enums";

import {
    Button,
    Checkbox,
    Field,
    FieldError,
    FieldGroup,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import {
    type CreateOrEditSslProviderFormInput,
    type CreateOrEditSslProviderFormOutput,
    CreateOrEditSslProviderFormSchema,
    SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
} from "./create-or-edit-ssl-provider.form.schema";

const DEFAULT_FORM_VALUES: CreateOrEditSslProviderFormInput = {
    name: "",
    kind: ESslProviderKind.LetsEncrypt,
    email: "",
    defaultKeyType: SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
    eabKid: "",
    eabHmacKey: "",
    availableInProjects: false,
    default: false,
};

export function CreateOrEditSslProviderForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    showAvailableInProjects,
    isEdit,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditSslProviderFormInput, unknown, CreateOrEditSslProviderFormOutput>({
        defaultValues: {
            ...DEFAULT_FORM_VALUES,
            ...initialValues,
        },
        resolver: zodResolver(CreateOrEditSslProviderFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        if (savedVersion === 0) {
            return;
        }

        reset(getValues());
        onHasChanges?.(false);
    }, [getValues, onHasChanges, reset, savedVersion]);

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, onHasChanges, isReadOnly]);

    const kind = useWatch({ control, name: "kind" });
    const showEabFields = kind === ESslProviderKind.ZeroSSL || kind === ESslProviderKind.GoogleTrust;

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kindField,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({ name: "email", control });
    const {
        field: defaultKeyType,
        fieldState: { invalid: isDefaultKeyTypeInvalid },
    } = useController({ name: "defaultKeyType", control });
    const {
        field: eabKid,
        fieldState: { invalid: isEabKidInvalid },
    } = useController({ name: "eabKid", control });
    const {
        field: eabHmacKey,
        fieldState: { invalid: isEabHmacKeyInvalid },
    } = useController({ name: "eabHmacKey", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditSslProviderFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditSslProviderFormOutput>) {
        console.error(_errors);
    }

    return (
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
                                    placeholder="my provider"
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
                                    value={kindField.value}
                                    onValueChange={value => {
                                        kindField.onChange(value);
                                    }}
                                    disabled={isReadOnly || isEdit}
                                >
                                    <SelectTrigger aria-invalid={isKindInvalid}>
                                        <SelectValue placeholder="Select provider" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {SSL_PROVIDER_OPTIONS.map(option => (
                                            <SelectItem
                                                key={option.value}
                                                value={option.value}
                                            >
                                                {option.label}
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
                                label="E-mail"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...email}
                                    type="email"
                                    placeholder="email@mail.com"
                                    aria-invalid={isEmailInvalid}
                                />
                                <FieldError errors={[errors.email]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Default Key Type" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Select
                                    value={defaultKeyType.value}
                                    onValueChange={value => {
                                        defaultKeyType.onChange(value);
                                    }}
                                    disabled={isReadOnly}
                                >
                                    <SelectTrigger aria-invalid={isDefaultKeyTypeInvalid}>
                                        <SelectValue placeholder="select key type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={SSL_PROVIDER_UNSPECIFIED_KEY_TYPE}>Unspecified</SelectItem>
                                        {SSL_KEY_TYPE_OPTIONS.map(option => (
                                            <SelectItem
                                                key={option.value}
                                                value={option.value}
                                            >
                                                {option.label}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FieldError errors={[errors.defaultKeyType]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    {showEabFields && (
                        <>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="EAB KID"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...eabKid}
                                            placeholder="key identifier"
                                            aria-invalid={isEabKidInvalid}
                                        />
                                        <FieldError errors={[errors.eabKid]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="EAB HMAC"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <PasswordInput
                                            value={eabHmacKey.value}
                                            onChange={eabHmacKey.onChange}
                                            placeholder="hmac"
                                            aria-invalid={isEabHmacKeyInvalid}
                                        />
                                        <FieldError errors={[errors.eabHmacKey]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </>
                    )}

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
                </fieldset>
            </div>
            {!isReadOnly && (
                <FormActionBar>
                    <SettingsFormCancelAction
                        onCancel={onClose}
                        disabled={isPending}
                    />
                    <Button
                        type="submit"
                        isLoading={isPending}
                        className="min-w-[100px]"
                    >
                        Save
                    </Button>
                </FormActionBar>
            )}
            {isReadOnly && (
                <FormActionBar>
                    <Button
                        type="button"
                        onClick={onClose}
                    >
                        Close
                    </Button>
                </FormActionBar>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditSslProviderFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditSslProviderFormInput>;
    showAvailableInProjects: boolean;
    isEdit: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
