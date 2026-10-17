import { useEffect, useId } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EOAuthKind } from "@application/shared/enums";
import { isPaidSignInKind } from "@application/shared/utils";

import { Button, Checkbox, Field, FieldDescription, FieldError, FieldGroup, Input } from "@/components/ui";

import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type { CreateOrEditOAuthFormInput, CreateOrEditOAuthFormOutput } from "./create-or-edit-oauth.form.schema";
import { CreateOrEditOAuthFormSchema, hasCustomURLs } from "./create-or-edit-oauth.form.schema";
import { OAUTH_KEPT_ON_BEE_NOTE, OAUTH_UNAVAILABLE_ON_BEE_NOTE, PaidPlanUpgradeLink } from "./oauth-on-bee.com";

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
    paidKindsUnavailable = false,
    pricingUrl,
    isActive = false,
    onClose,
}: Props) {
    const isReadOnly = readOnly;
    const paidKindsNoteId = useId();

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
            trustEmail: initialValues?.trustEmail ?? false,
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
    const showTrustEmail = currentProvider === EOAuthKind.OpenIDConnect;
    const showCustomURLs = hasCustomURLs(currentProvider);
    // On Bee the paid kinds can't be picked, wherever a kind is picked. One that
    // exists stays editable - its secret may need replacing - and says what Bee
    // leaves it.
    const marksPaidKinds = paidKindsUnavailable && !disableProvider;
    const isPaidProviderOnBee = paidKindsUnavailable && disableProvider && isPaidSignInKind(currentProvider);

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
    const { field: trustEmail } = useController({ name: "trustEmail", control });
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
                {isPaidProviderOnBee && (
                    <div className={cn(dashedBorderBox, "mb-4")}>
                        {isActive ? OAUTH_KEPT_ON_BEE_NOTE : OAUTH_UNAVAILABLE_ON_BEE_NOTE}{" "}
                        <PaidPlanUpgradeLink pricingUrl={pricingUrl} />
                    </div>
                )}
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
                                    <SelectTrigger
                                        aria-invalid={isKindInvalid}
                                        aria-describedby={marksPaidKinds ? paidKindsNoteId : undefined}
                                    >
                                        <SelectValue placeholder="Select provider" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {providerOptions.map(option => {
                                            const isUnavailable = marksPaidKinds && isPaidSignInKind(option.value);

                                            return (
                                                <SelectItem
                                                    key={option.value}
                                                    value={option.value}
                                                    disabled={isUnavailable}
                                                >
                                                    {isUnavailable ? `${option.label} · paid plan` : option.label}
                                                </SelectItem>
                                            );
                                        })}
                                    </SelectContent>
                                </Select>
                                {marksPaidKinds && (
                                    <FieldDescription id={paidKindsNoteId}>
                                        Google, Microsoft Online and OpenID Connect sign-in need a paid plan: this
                                        installation is on Bee. <PaidPlanUpgradeLink pricingUrl={pricingUrl} />
                                    </FieldDescription>
                                )}
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

                    {showCustomURLs && (
                        <>
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
                            <FieldDescription className="text-xs">
                                For a GitLab or Gitea of your own, all three. Leave them empty for gitlab.com or
                                gitea.com.
                            </FieldDescription>
                        </>
                    )}

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

                    {showTrustEmail && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Trust emails" />}
                        >
                            <Field>
                                <div className="flex items-center gap-3">
                                    <Checkbox
                                        id="oauth-trust-email"
                                        checked={trustEmail.value}
                                        onCheckedChange={checked => {
                                            trustEmail.onChange(Boolean(checked));
                                        }}
                                    />
                                    <label
                                        htmlFor="oauth-trust-email"
                                        className="text-sm"
                                    >
                                        Take every email this provider gives as verified
                                    </label>
                                </div>
                                <FieldDescription className="text-xs">
                                    A sign-in is matched to a user by email, so it is refused unless the provider says
                                    the email is verified (email_verified). Turn this on only for an identity provider
                                    whose users and emails you manage: with one anybody can sign up to, whoever types in
                                    a user&apos;s email could sign in as them.
                                </FieldDescription>
                            </Field>
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
                <FormActionBar contentClassName="gap-4">
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
                        className="min-w-[100px]"
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
    onSubmit: (values: CreateOrEditOAuthFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditOAuthFormInput>;
    disableProvider: boolean;
    readOnly?: boolean;
    /**
     * The plan in force allows no Google, Microsoft Online or OpenID Connect
     * provider to be added or enabled: Bee. False when that is not known.
     */
    paidKindsUnavailable?: boolean;
    /** Where a paid plan is bought, for an Upgrade link. No link when empty. */
    pricingUrl?: string;
    /** The provider being edited is active. */
    isActive?: boolean;
    onClose?: () => void;
}
