import { useEffect, useMemo } from "react";

import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { ProjectAcmeDnsProviderQueries, ProjectSslProviderQueries } from "~/projects/data/queries";
import { AcmeDnsProviderQueries, SslProviderQueries } from "~/settings/data/queries";
import {
    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
    SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS,
} from "~/settings/module-shared/constants/settings-form-layout.constants";
import { SSL_CERT_TYPE_OPTIONS } from "~/settings/module-shared/constants/ssl-provider.constants";
import { useNotificationSettingsSources } from "~/settings/module-shared/hooks";

import {
    AppLink,
    Combobox,
    ContentBlock,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ESslCertType, ESslKeyType, ESslProviderKind } from "@application/shared/enums";
import { NotificationSettings } from "@application/shared/form";

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
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";
import type { SslCertTableScope } from "../ssl-cert-table";

import type {
    CreateOrEditSslCertFormInput,
    CreateOrEditSslCertFormOutput,
} from "./create-or-edit-ssl-cert.form.schema";
import { CreateOrEditSslCertFormSchema } from "./create-or-edit-ssl-cert.form.schema";

const LETS_ENCRYPT_KEY_TYPES: ESslKeyType[] = [
    ESslKeyType.ECP256,
    ESslKeyType.ECP384,
    ESslKeyType.RSA2048,
    ESslKeyType.RSA3072,
    ESslKeyType.RSA4096,
];

const CUSTOM_KEY_TYPES: ESslKeyType[] = [
    ESslKeyType.ECP256,
    ESslKeyType.ECP384,
    ESslKeyType.ECP521,
    ESslKeyType.RSA2048,
    ESslKeyType.RSA3072,
    ESslKeyType.RSA4096,
];

type SettingProviderOption = Record<"id" | "name", string>;

function getProviderKind(certType: ESslCertType): ESslProviderKind | undefined {
    switch (certType) {
        case ESslCertType.LetsEncrypt:
            return ESslProviderKind.LetsEncrypt;
        case ESslCertType.ZeroSSL:
            return ESslProviderKind.ZeroSSL;
        case ESslCertType.GoogleTrust:
            return ESslProviderKind.GoogleTrust;
        default:
            return undefined;
    }
}

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

function addDays(date: Date, days: number): Date {
    const result = new Date(date);
    result.setDate(result.getDate() + days);
    return result;
}

export function CreateOrEditSslCertForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    scope,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;
    const initialProviderId = initialValues?.provider?.id;
    const initialProviderName = initialValues?.provider?.name;
    const initialProvider = useMemo(
        () => (initialProviderId ? { id: initialProviderId, name: initialProviderName ?? "" } : undefined),
        [initialProviderId, initialProviderName],
    );
    const initialAcmeProviderId = initialValues?.acmeProvider?.id;
    const initialAcmeProviderName = initialValues?.acmeProvider?.name;
    const initialAcmeProvider = useMemo(
        () => (initialAcmeProviderId ? { id: initialAcmeProviderId, name: initialAcmeProviderName ?? "" } : undefined),
        [initialAcmeProviderId, initialAcmeProviderName],
    );

    const methods = useForm<CreateOrEditSslCertFormInput, unknown, CreateOrEditSslCertFormOutput>({
        defaultValues: {
            domain: initialValues?.domain ?? "",
            certType: initialValues?.certType ?? ESslCertType.LetsEncrypt,
            provider: initialProvider,
            acmeProvider: initialAcmeProvider,
            email: initialValues?.email ?? "",
            keyType: initialValues?.keyType ?? ESslKeyType.ECP256,
            autoRenew: initialValues?.autoRenew ?? true,
            certificate: initialValues?.certificate ?? "",
            privateKey: initialValues?.privateKey ?? "",
            expireAt: initialValues?.expireAt ?? null,
            notifyFrom: initialValues?.notifyFrom ?? null,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
            notification: {
                successUseDefault: initialValues?.notification?.successUseDefault ?? true,
                success: initialValues?.notification?.success ?? undefined,
                failureUseDefault: initialValues?.notification?.failureUseDefault ?? true,
                failure: initialValues?.notification?.failure ?? undefined,
            },
        },
        resolver: zodResolver(CreateOrEditSslCertFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
        getValues,
        reset,
        setValue,
    } = methods;

    useEffect(() => {
        reset({
            domain: initialValues?.domain ?? "",
            certType: initialValues?.certType ?? ESslCertType.LetsEncrypt,
            provider: initialProvider,
            acmeProvider: initialAcmeProvider,
            email: initialValues?.email ?? "",
            keyType: initialValues?.keyType ?? ESslKeyType.ECP256,
            autoRenew: initialValues?.autoRenew ?? true,
            certificate: initialValues?.certificate ?? "",
            privateKey: initialValues?.privateKey ?? "",
            expireAt: initialValues?.expireAt ?? null,
            notifyFrom: initialValues?.notifyFrom ?? null,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
            notification: {
                successUseDefault: initialValues?.notification?.successUseDefault ?? true,
                success: initialValues?.notification?.success ?? undefined,
                failureUseDefault: initialValues?.notification?.failureUseDefault ?? true,
                failure: initialValues?.notification?.failure ?? undefined,
            },
        });
    }, [
        initialValues?.autoRenew,
        initialValues?.availableInProjects,
        initialValues?.certType,
        initialValues?.certificate,
        initialValues?.default,
        initialValues?.domain,
        initialValues?.email,
        initialValues?.expireAt,
        initialValues?.keyType,
        initialValues?.notification?.failure,
        initialValues?.notification?.failureUseDefault,
        initialValues?.notification?.success,
        initialValues?.notification?.successUseDefault,
        initialValues?.notifyFrom,
        initialValues?.privateKey,
        initialAcmeProvider,
        initialProvider,
        reset,
    ]);

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

    const certType = useWatch({ control, name: "certType" });
    const domainValue = useWatch({ control, name: "domain" });
    const providerValue = useWatch({ control, name: "provider" });
    const acmeProviderValue = useWatch({ control, name: "acmeProvider" });
    const expireAt = useWatch({ control, name: "expireAt" });
    const notifyFrom = useWatch({ control, name: "notifyFrom" });
    const providerKind = getProviderKind(certType);
    const isCustom = certType === ESslCertType.Custom;
    const isAcme = providerKind !== undefined;
    const requiresProvider = certType === ESslCertType.ZeroSSL || certType === ESslCertType.GoogleTrust;
    const isWildcardAcme = isAcme && domainValue.trim().startsWith("*.");
    const projectId = scope.type === "project" ? scope.projectId : "";
    const sslProvidersManageRoute =
        scope.type === "project"
            ? ROUTE.projects.single.providerConfiguration.sslProviders.$route(scope.projectId)
            : ROUTE.settings.sslProviders.$route;
    const acmeDnsProvidersManageRoute =
        scope.type === "project"
            ? ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$route(scope.projectId)
            : ROUTE.settings.acmeDnsProviders.$route;
    const { sources: notificationSources, manageLink: notificationManageLink } = useNotificationSettingsSources(scope);

    const settingsProviderQuery = SslProviderQueries.useFindManyPaginated(
        { kind: providerKind },
        {
            enabled: providerKind !== undefined && scope.type === "settings",
        },
    );

    const projectProviderQuery = ProjectSslProviderQueries.useFindManyPaginated(
        {
            projectID: projectId,
            kind: providerKind,
        },
        {
            enabled: providerKind !== undefined && scope.type === "project" && projectId.length > 0,
        },
    );

    const providerQuery = scope.type === "project" ? projectProviderQuery : settingsProviderQuery;
    const providerOptions = useMemo(() => providerQuery.data?.data ?? [], [providerQuery.data?.data]);
    const providerComboboxOptions = useMemo(
        () =>
            providerOptions.map(option => ({
                value: { id: option.id, name: option.name } satisfies SettingProviderOption,
                label: option.name,
            })),
        [providerOptions],
    );

    const settingsAcmeProviderQuery = AcmeDnsProviderQueries.useFindManyPaginated(
        {},
        {
            enabled: isAcme && scope.type === "settings",
        },
    );

    const projectAcmeProviderQuery = ProjectAcmeDnsProviderQueries.useFindManyPaginated(
        {
            projectID: projectId,
        },
        {
            enabled: isAcme && scope.type === "project" && projectId.length > 0,
        },
    );

    const acmeProviderQuery = scope.type === "project" ? projectAcmeProviderQuery : settingsAcmeProviderQuery;
    const acmeProviderOptions = useMemo(() => acmeProviderQuery.data?.data ?? [], [acmeProviderQuery.data?.data]);
    const acmeProviderComboboxOptions = useMemo(
        () =>
            acmeProviderOptions.map(option => ({
                value: { id: option.id, name: option.name } satisfies SettingProviderOption,
                label: option.name,
            })),
        [acmeProviderOptions],
    );

    useEffect(() => {
        if (isAcme) {
            setValue("keyType", ESslKeyType.ECP256);
        }
    }, [isAcme, setValue]);

    useEffect(() => {
        if (providerKind !== undefined || !providerValue) {
            return;
        }

        setValue("provider", undefined, { shouldDirty: true });
    }, [providerKind, providerValue, setValue]);

    useEffect(() => {
        if (providerKind === undefined || !providerValue || providerQuery.isFetching || providerOptions.length === 0) {
            return;
        }

        if (!providerOptions.some(option => option.id === providerValue.id)) {
            setValue("provider", undefined, { shouldDirty: true });
        }
    }, [providerKind, providerOptions, providerQuery.isFetching, providerValue, setValue]);

    useEffect(() => {
        if (isAcme || !acmeProviderValue) {
            return;
        }

        setValue("acmeProvider", undefined, { shouldDirty: true });
    }, [acmeProviderValue, isAcme, setValue]);

    useEffect(() => {
        if (!isAcme || !acmeProviderValue || acmeProviderQuery.isFetching || acmeProviderOptions.length === 0) {
            return;
        }

        if (!acmeProviderOptions.some(option => option.id === acmeProviderValue.id)) {
            setValue("acmeProvider", undefined, { shouldDirty: true });
        }
    }, [acmeProviderOptions, acmeProviderQuery.isFetching, acmeProviderValue, isAcme, setValue]);

    useEffect(() => {
        if (!expireAt || notifyFrom) {
            return;
        }

        setValue("notifyFrom", addDays(expireAt, -30), { shouldDirty: true });
    }, [expireAt, notifyFrom, setValue]);

    const keyTypeOptions = useMemo(() => {
        return (isAcme ? LETS_ENCRYPT_KEY_TYPES : CUSTOM_KEY_TYPES).map(value => ({
            value,
            label: formatKeyTypeLabel(value),
        }));
    }, [isAcme]);

    const {
        field: domain,
        fieldState: { invalid: isDomainInvalid },
    } = useController({ name: "domain", control });
    const { field: certTypeField } = useController({ name: "certType", control });
    const {
        field: provider,
        fieldState: { invalid: isProviderInvalid },
    } = useController({ name: "provider", control });
    const {
        field: acmeProvider,
        fieldState: { invalid: isAcmeProviderInvalid },
    } = useController({ name: "acmeProvider", control });
    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({ name: "email", control });
    const {
        field: keyType,
        fieldState: { invalid: isKeyTypeInvalid },
    } = useController({ name: "keyType", control });
    const { field: autoRenew } = useController({ name: "autoRenew", control });
    const {
        field: certificate,
        fieldState: { invalid: isCertificateInvalid },
    } = useController({ name: "certificate", control });
    const {
        field: privateKey,
        fieldState: { invalid: isPrivateKeyInvalid },
    } = useController({ name: "privateKey", control });
    const {
        field: expireAtField,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const {
        field: notifyFromField,
        fieldState: { invalid: isNotifyFromInvalid },
    } = useController({ name: "notifyFrom", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditSslCertFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditSslCertFormOutput>) {
        console.error(_errors);
    }

    return (
        <FormProvider {...methods}>
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
                        <FieldGroup>
                            <InfoBlock
                                titleWidth={220}
                                title={<LabelWithInfo label="Domain" />}
                            >
                                <Field>
                                    <Input
                                        {...domain}
                                        aria-invalid={isDomainInvalid}
                                    />
                                    <FieldError errors={[errors.domain]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Certificate Type"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Select
                                        value={certTypeField.value}
                                        onValueChange={value => {
                                            certTypeField.onChange(value);
                                        }}
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="select certificate type" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {SSL_CERT_TYPE_OPTIONS.map(option => (
                                                <SelectItem
                                                    key={option.value}
                                                    value={option.value}
                                                >
                                                    {option.label}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <FieldError errors={[errors.certType]} />
                                </Field>
                            </InfoBlock>

                            {providerKind !== undefined && (
                                <>
                                    <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                                        <span className="text-orange-500">Note:</span> SSL provider is required if you
                                        select <span className="text-orange-500">Zero SSL</span> or{" "}
                                        <span className="text-orange-500">Google Trust Services</span> as the
                                        certificate type.
                                    </div>

                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="SSL Provider"
                                                isRequired={requiresProvider}
                                            />
                                        }
                                    >
                                        <Field>
                                            <Combobox<SettingProviderOption>
                                                options={providerComboboxOptions}
                                                value={provider.value?.id ?? null}
                                                onChange={(_, option) => {
                                                    provider.onChange(option ?? undefined);
                                                }}
                                                placeholder="select provider"
                                                searchable
                                                allowClear
                                                closeOnSelect
                                                emptyText="No SSL providers available"
                                                className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                valueKey="id"
                                                aria-invalid={isProviderInvalid}
                                                loading={providerQuery.isFetching}
                                                onRefresh={() => void providerQuery.refetch()}
                                                isRefreshing={providerQuery.isRefetching}
                                                disabled={isReadOnly}
                                            />
                                            <FieldError errors={[errors.provider]} />
                                            <AppLink.Modules
                                                to={sslProvidersManageRoute}
                                                className="text-sm text-link"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                ignorePrevPath
                                            >
                                                Configure SSL providers
                                            </AppLink.Modules>
                                        </Field>
                                    </InfoBlock>
                                </>
                            )}

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label={isCustom ? "E-mail" : "Registration E-mail"}
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Input
                                        {...email}
                                        type="email"
                                        aria-invalid={isEmailInvalid}
                                    />
                                    <FieldError errors={[errors.email]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Key Type"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Select
                                        value={keyType.value}
                                        onValueChange={value => {
                                            keyType.onChange(value);
                                        }}
                                    >
                                        <SelectTrigger aria-invalid={isKeyTypeInvalid}>
                                            <SelectValue placeholder="Select key type" />
                                        </SelectTrigger>
                                        <SelectContent>
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
                                    <FieldError errors={[errors.keyType]} />
                                </Field>
                            </InfoBlock>

                            {isAcme && (
                                <>
                                    <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                                        <span className="text-orange-500">Note:</span> ACME DNS provider is required if
                                        your domain is a wildcard domain.
                                    </div>

                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="ACME DNS Provider"
                                                isRequired={isWildcardAcme}
                                            />
                                        }
                                    >
                                        <Field>
                                            <Combobox<SettingProviderOption>
                                                options={acmeProviderComboboxOptions}
                                                value={acmeProvider.value?.id ?? null}
                                                onChange={(_, option) => {
                                                    acmeProvider.onChange(option ?? undefined);
                                                }}
                                                placeholder="select ACME DNS provider"
                                                searchable
                                                allowClear
                                                closeOnSelect
                                                emptyText="No ACME DNS providers available"
                                                className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                valueKey="id"
                                                aria-invalid={isAcmeProviderInvalid}
                                                loading={acmeProviderQuery.isFetching}
                                                onRefresh={() => void acmeProviderQuery.refetch()}
                                                isRefreshing={acmeProviderQuery.isRefetching}
                                                disabled={isReadOnly}
                                            />
                                            <FieldError errors={[errors.acmeProvider]} />
                                            <AppLink.Modules
                                                to={acmeDnsProvidersManageRoute}
                                                className="text-sm text-link"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                ignorePrevPath
                                            >
                                                Configure ACME DNS providers
                                            </AppLink.Modules>
                                        </Field>
                                    </InfoBlock>
                                </>
                            )}

                            {!isCustom ? (
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Auto-renew" />}
                                >
                                    <Checkbox
                                        checked={autoRenew.value}
                                        onCheckedChange={checked => {
                                            autoRenew.onChange(Boolean(checked));
                                        }}
                                    />
                                </InfoBlock>
                            ) : (
                                <>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Certificate"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <Textarea
                                                {...certificate}
                                                aria-invalid={isCertificateInvalid}
                                                rows={4}
                                                maxRows={7}
                                            />
                                            <FieldError errors={[errors.certificate]} />
                                        </Field>
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
                                        <Field>
                                            <Textarea
                                                {...privateKey}
                                                aria-invalid={isPrivateKeyInvalid}
                                                rows={4}
                                                maxRows={7}
                                            />
                                            <FieldError errors={[errors.privateKey]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Expire At"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                            <DateTimePicker
                                                value={expireAtField.value ?? undefined}
                                                onChange={date => {
                                                    expireAtField.onChange(date ?? null);
                                                }}
                                                displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                                granularity="second"
                                                showClearButton
                                                aria-invalid={isExpireAtInvalid}
                                            />
                                            <FieldError errors={[errors.expireAt]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Notify From" />}
                                    >
                                        <Field className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                            <DateTimePicker
                                                value={notifyFromField.value ?? undefined}
                                                onChange={date => {
                                                    notifyFromField.onChange(date ?? null);
                                                }}
                                                displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                                granularity="second"
                                                showClearButton
                                                aria-invalid={isNotifyFromInvalid}
                                            />
                                            <FieldError errors={[errors.notifyFrom]} />
                                        </Field>
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
                        </FieldGroup>

                        <ContentBlock label="Notification Configuration">
                            <NotificationSettings<CreateOrEditSslCertFormInput>
                                names={{
                                    successUseDefault: "notification.successUseDefault",
                                    success: "notification.success",
                                    failureUseDefault: "notification.failureUseDefault",
                                    failure: "notification.failure",
                                }}
                                sources={notificationSources}
                                manageLink={notificationManageLink}
                                readOnly={isReadOnly}
                                titleWidth={220}
                            />
                        </ContentBlock>
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
        </FormProvider>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditSslCertFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditSslCertFormInput>;
    scope: SslCertTableScope;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
