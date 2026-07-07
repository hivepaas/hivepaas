import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";

import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";
import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EOAuthKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type { CreateOrEditOAuthFormInput, CreateOrEditOAuthFormOutput } from "./create-or-edit-oauth.form.schema";
import { CreateOrEditOAuthFormSchema } from "./create-or-edit-oauth.form.schema";

const providerOptions = [
    { value: EOAuthKind.Github, label: "Github" },
    { value: EOAuthKind.Gitlab, label: "Gitlab" },
    { value: EOAuthKind.Gitea, label: "Gitea" },
    { value: EOAuthKind.Google, label: "Google" },
    { value: EOAuthKind.MicrosoftOnline, label: "Microsoft Online" },
    { value: EOAuthKind.OpenIDConnect, label: "OpenID Connect" },
];

export function CreateOrEditOAuthForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    disableProvider,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnly;

    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditOAuthFormInput, unknown, CreateOrEditOAuthFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? EOAuthKind.Github,
            organization: initialValues?.organization ?? "",
            clientId: initialValues?.clientId ?? "",
            clientSecret: initialValues?.clientSecret ?? "",
            authURL: initialValues?.authURL ?? "",
            tokenURL: initialValues?.tokenURL ?? "",
            profileURL: initialValues?.profileURL ?? "",
            autoDiscoveryURL: initialValues?.autoDiscoveryURL ?? "",
            scopes: initialValues?.scopes ?? "",
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditOAuthFormSchema),
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
    }, [isDirty, isReadOnly, onHasChanges]);

    const currentProvider = useWatch({ control, name: "kind" });
    const showAutoDiscoveryURL = currentProvider === EOAuthKind.OpenIDConnect;

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: organization,
        fieldState: { invalid: isOrganizationInvalid },
    } = useController({ name: "organization", control });
    const {
        field: clientId,
        fieldState: { invalid: isClientIdInvalid },
    } = useController({ name: "clientId", control });
    const {
        field: clientSecret,
        fieldState: { invalid: isClientSecretInvalid },
    } = useController({ name: "clientSecret", control });
    const {
        field: authURL,
        fieldState: { invalid: isAuthURLInvalid },
    } = useController({ name: "authURL", control });
    const {
        field: tokenURL,
        fieldState: { invalid: isTokenURLInvalid },
    } = useController({ name: "tokenURL", control });
    const {
        field: profileURL,
        fieldState: { invalid: isProfileURLInvalid },
    } = useController({ name: "profileURL", control });
    const {
        field: autoDiscoveryURL,
        fieldState: { invalid: isAutoDiscoveryURLInvalid },
    } = useController({ name: "autoDiscoveryURL", control });
    const {
        field: scopes,
        fieldState: { invalid: isScopesInvalid },
    } = useController({ name: "scopes", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditOAuthFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditOAuthFormOutput>) {
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
                {isReadOnly && <PermissionReadonlyNotice />}
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
                                    disabled={disableProvider || isReadOnly}
                                >
                                    <SelectTrigger aria-invalid={isKindInvalid}>
                                        <SelectValue placeholder="Select provider" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {providerOptions.map(option => (
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
                                label="Organization"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...organization}
                                    aria-invalid={isOrganizationInvalid}
                                />
                                <FieldError errors={[errors.organization]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Client ID"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...clientId}
                                    aria-invalid={isClientIdInvalid}
                                />
                                <FieldError errors={[errors.clientId]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Client Secret"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <PasswordInput
                                    value={clientSecret.value}
                                    onChange={clientSecret.onChange}
                                    aria-invalid={isClientSecretInvalid}
                                />
                                <FieldError errors={[errors.clientSecret]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Auth URL" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...authURL}
                                    aria-invalid={isAuthURLInvalid}
                                />
                                <FieldError errors={[errors.authURL]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Token URL" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...tokenURL}
                                    aria-invalid={isTokenURLInvalid}
                                />
                                <FieldError errors={[errors.tokenURL]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Profile URL" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...profileURL}
                                    aria-invalid={isProfileURLInvalid}
                                />
                                <FieldError errors={[errors.profileURL]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    {showAutoDiscoveryURL && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Auto-Discovery URL" />}
                        >
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...autoDiscoveryURL}
                                        aria-invalid={isAutoDiscoveryURLInvalid}
                                    />
                                    <FieldError errors={[errors.autoDiscoveryURL]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    )}

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Scopes" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...scopes}
                                    placeholder="email, profile, ..."
                                    aria-invalid={isScopesInvalid}
                                />
                                <FieldError errors={[errors.scopes]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

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
                <div className="shrink-0 px-0 mt-6 pb-6">
                    <div className="flex items-center justify-end gap-4">
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
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditOAuthFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditOAuthFormInput>;
    disableProvider: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
