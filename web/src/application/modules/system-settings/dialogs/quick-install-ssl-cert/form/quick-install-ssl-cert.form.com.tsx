import React, { useEffect, useMemo } from "react";

import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@components/ui/dialog";
import { Field, FieldError, FieldGroup, FieldLabel } from "@components/ui/field";
import { Input } from "@components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useController, useForm, useWatch } from "react-hook-form";
import { AcmeDnsProviderQueries, SslProviderQueries } from "~/settings/data/queries";
import { SSL_CERT_TYPE_OPTIONS } from "~/settings/module-shared/constants/ssl-provider.constants";

import { AppLink, Combobox, DatePicker, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ESslCertType, ESslKeyType, ESslProviderKind } from "@application/shared/enums";

import {
    type QuickInstallSslCertFormInput,
    type QuickInstallSslCertFormOutput,
    QuickInstallSslCertFormSchema,
} from "../schemas";

const ACME_KEY_TYPES: ESslKeyType[] = [
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

type ProviderOption = Record<"id" | "name", string>;

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

function toWildcardDomain(domain: string): string {
    const parts = domain.split(".");
    return `*.${parts.length > 2 ? parts.slice(1).join(".") : domain}`;
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

export function QuickInstallSslCertForm({
    domain,
    isPending,
    prefill,
    readOnly = false,
    onSubmit,
    onHasChanges,
}: Props) {
    const defaultEmail = "";
    const defaultKeyType = ESslKeyType.ECP256;
    const defaultAutoRenew = true;
    const prefillEmail = prefill?.email ?? defaultEmail;
    const prefillKeyType = prefill?.keyType ?? defaultKeyType;
    const prefillAutoRenew = prefill?.autoRenew ?? defaultAutoRenew;

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
        setValue,
        reset,
    } = useForm<QuickInstallSslCertFormInput, unknown, QuickInstallSslCertFormOutput>({
        defaultValues: {
            name: domain,
            domain,
            wildcardDomain: false,
            certType: ESslCertType.LetsEncrypt,
            provider: undefined,
            acmeProvider: undefined,
            email: prefillEmail,
            keyType: prefillKeyType,
            autoRenew: prefillAutoRenew,
            certificate: "",
            privateKey: "",
            expireAt: null,
            notifyFrom: null,
        },
        resolver: zodResolver(QuickInstallSslCertFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        reset({
            name: domain,
            domain,
            wildcardDomain: false,
            certType: ESslCertType.LetsEncrypt,
            provider: undefined,
            acmeProvider: undefined,
            email: prefillEmail,
            keyType: prefillKeyType,
            autoRenew: prefillAutoRenew,
            certificate: "",
            privateKey: "",
            expireAt: null,
            notifyFrom: null,
        });
    }, [domain, prefillAutoRenew, prefillEmail, prefillKeyType, reset]);

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    const certType = useWatch({ control, name: "certType" });
    const providerValue = useWatch({ control, name: "provider" });
    const acmeProviderValue = useWatch({ control, name: "acmeProvider" });
    const expireAt = useWatch({ control, name: "expireAt" });
    const notifyFrom = useWatch({ control, name: "notifyFrom" });
    const wildcardDomainWatched = useWatch({ control, name: "wildcardDomain" });
    const effectiveDomain = wildcardDomainWatched ? toWildcardDomain(domain) : domain;

    const isCustom = certType === ESslCertType.Custom;
    const isAcme = !isCustom && certType !== ESslCertType.SelfSigned;
    const requiresProvider = certType === ESslCertType.ZeroSSL || certType === ESslCertType.GoogleTrust;
    const providerKind = getProviderKind(certType);

    const sslProvidersRoute = ROUTE.settings.sslProviders.$route;
    const acmeDnsProvidersRoute = ROUTE.settings.acmeDnsProviders.$route;

    const providerQuery = SslProviderQueries.useFindManyPaginated({ kind: providerKind }, { enabled: isAcme });
    const acmeProviderQuery = AcmeDnsProviderQueries.useFindManyPaginated({}, { enabled: isAcme });

    const providerOptions = useMemo(() => providerQuery.data?.data ?? [], [providerQuery.data?.data]);
    const providerComboboxOptions = useMemo(
        () =>
            providerOptions.map(option => ({
                value: { id: option.id, name: option.name } satisfies ProviderOption,
                label: option.name,
            })),
        [providerOptions],
    );

    const acmeProviderOptions = useMemo(() => acmeProviderQuery.data?.data ?? [], [acmeProviderQuery.data?.data]);
    const acmeProviderComboboxOptions = useMemo(
        () =>
            acmeProviderOptions.map(option => ({
                value: { id: option.id, name: option.name } satisfies ProviderOption,
                label: option.name,
            })),
        [acmeProviderOptions],
    );

    // Reset key type to ECP256 when switching to ACME types
    useEffect(() => {
        if (isAcme) {
            setValue("keyType", ESslKeyType.ECP256);
        }
    }, [isAcme, setValue]);

    // Clear provider when cert type changes away from provider-required types
    useEffect(() => {
        if (providerKind !== undefined || !providerValue) {
            return;
        }
        setValue("provider", undefined, { shouldDirty: true });
    }, [providerKind, providerValue, setValue]);

    // Clear provider if no longer in available options
    useEffect(() => {
        if (providerKind === undefined || !providerValue || providerQuery.isFetching || providerOptions.length === 0) {
            return;
        }
        if (!providerOptions.some(o => o.id === providerValue.id)) {
            setValue("provider", undefined, { shouldDirty: true });
        }
    }, [providerKind, providerOptions, providerQuery.isFetching, providerValue, setValue]);

    // Clear ACME provider when switching to Custom
    useEffect(() => {
        if (isAcme || !acmeProviderValue) {
            return;
        }
        setValue("acmeProvider", undefined, { shouldDirty: true });
    }, [acmeProviderValue, isAcme, setValue]);

    // Clear ACME provider if no longer in available options
    useEffect(() => {
        if (!isAcme || !acmeProviderValue || acmeProviderQuery.isFetching || acmeProviderOptions.length === 0) {
            return;
        }
        if (!acmeProviderOptions.some(o => o.id === acmeProviderValue.id)) {
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
        return (isCustom ? CUSTOM_KEY_TYPES : ACME_KEY_TYPES).map(value => ({
            value,
            label: formatKeyTypeLabel(value),
        }));
    }, [isCustom]);

    const { field: certTypeField } = useController({ name: "certType", control });
    const { field: wildcardDomain } = useController({ name: "wildcardDomain", control });
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
    const { field: keyType } = useController({ name: "keyType", control });
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

    function onValid(values: QuickInstallSslCertFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                if (readOnly) {
                    return;
                }

                void handleSubmit(onValid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <fieldset
                disabled={readOnly}
                className="contents"
            >
                <DialogBody>
                    <FieldGroup>
                        <Field>
                            <FieldLabel>Domain</FieldLabel>
                            <Input
                                value={effectiveDomain}
                                disabled
                            />
                        </Field>

                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="Wildcard Domain" />}
                                titleWidth={150}
                            >
                                <Checkbox
                                    checked={wildcardDomain.value}
                                    onCheckedChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        wildcardDomain.onChange(value === true);
                                    }}
                                    disabled={readOnly}
                                />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Certificate Type"
                                        isRequired
                                    />
                                }
                                titleWidth={150}
                            >
                                <Select
                                    value={certTypeField.value}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        certTypeField.onChange(value);
                                    }}
                                    disabled={readOnly}
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
                            </InfoBlock>
                        </Field>

                        {isAcme && (
                            <>
                                <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                                    <span className="text-orange-500">Note:</span> SSL provider is required if you
                                    select{" "}
                                    <AppLink.Modules
                                        to={sslProvidersRoute}
                                        className="text-link"
                                        ignorePrevPath
                                    >
                                        Zero SSL
                                    </AppLink.Modules>{" "}
                                    or{" "}
                                    <AppLink.Modules
                                        to={sslProvidersRoute}
                                        className="text-link"
                                        ignorePrevPath
                                    >
                                        Google Trust Services
                                    </AppLink.Modules>{" "}
                                    as the certificate type.
                                </div>

                                <Field>
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="SSL Provider"
                                                isRequired={requiresProvider}
                                            />
                                        }
                                        titleWidth={150}
                                    >
                                        <Combobox<ProviderOption>
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
                                            valueKey="id"
                                            aria-invalid={isProviderInvalid}
                                            loading={providerQuery.isFetching}
                                            onRefresh={() => void providerQuery.refetch()}
                                            isRefreshing={providerQuery.isRefetching}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.provider]} />
                                        <AppLink.Modules
                                            to={sslProvidersRoute}
                                            className="text-sm text-link"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            ignorePrevPath
                                        >
                                            Configure SSL providers
                                        </AppLink.Modules>
                                    </InfoBlock>
                                </Field>

                                <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                                    <span className="text-orange-500">Note:</span> ACME DNS provider is required if your
                                    domain is a wildcard domain.
                                </div>

                                <Field>
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="ACME DNS Provider"
                                                isRequired={wildcardDomain.value}
                                            />
                                        }
                                        titleWidth={150}
                                    >
                                        <Combobox<ProviderOption>
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
                                            valueKey="id"
                                            aria-invalid={isAcmeProviderInvalid}
                                            loading={acmeProviderQuery.isFetching}
                                            onRefresh={() => void acmeProviderQuery.refetch()}
                                            isRefreshing={acmeProviderQuery.isRefetching}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.acmeProvider]} />
                                        <AppLink.Modules
                                            to={acmeDnsProvidersRoute}
                                            className="text-sm text-link"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            ignorePrevPath
                                        >
                                            Configure ACME DNS providers
                                        </AppLink.Modules>
                                    </InfoBlock>
                                </Field>
                            </>
                        )}

                        <Field>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label={isCustom ? "E-mail" : "Registration E-mail"}
                                        isRequired
                                    />
                                }
                                titleWidth={150}
                            >
                                <Input
                                    {...email}
                                    type="email"
                                    aria-invalid={isEmailInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.email]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Key Type"
                                        isRequired
                                    />
                                }
                                titleWidth={150}
                            >
                                <Select
                                    value={keyType.value}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        keyType.onChange(value);
                                    }}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger>
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
                            </InfoBlock>
                        </Field>

                        {!isCustom ? (
                            <Field>
                                <InfoBlock
                                    title={
                                        <LabelWithInfo
                                            label="Auto-renew"
                                            isRequired
                                        />
                                    }
                                    titleWidth={150}
                                >
                                    <Checkbox
                                        checked={autoRenew.value}
                                        onCheckedChange={value => {
                                            if (readOnly) {
                                                return;
                                            }

                                            autoRenew.onChange(value === true);
                                        }}
                                        disabled={readOnly}
                                    />
                                </InfoBlock>
                            </Field>
                        ) : (
                            <>
                                <Field>
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="Certificate"
                                                isRequired
                                            />
                                        }
                                        titleWidth={150}
                                    >
                                        <Textarea
                                            {...certificate}
                                            aria-invalid={isCertificateInvalid}
                                            rows={4}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.certificate]} />
                                    </InfoBlock>
                                </Field>

                                <Field>
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="Private Key"
                                                isRequired
                                            />
                                        }
                                        titleWidth={150}
                                    >
                                        <Textarea
                                            {...privateKey}
                                            aria-invalid={isPrivateKeyInvalid}
                                            rows={4}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.privateKey]} />
                                    </InfoBlock>
                                </Field>

                                <Field>
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="Expire At"
                                                isRequired
                                            />
                                        }
                                        titleWidth={150}
                                    >
                                        <DatePicker
                                            value={expireAtField.value ?? undefined}
                                            onChange={date => {
                                                if (readOnly) {
                                                    return;
                                                }

                                                expireAtField.onChange(date ?? null);
                                            }}
                                            aria-invalid={isExpireAtInvalid}
                                            placeholder="Select date"
                                            allowClear
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.expireAt]} />
                                    </InfoBlock>
                                </Field>

                                <Field>
                                    <InfoBlock
                                        title={<LabelWithInfo label="Notify From" />}
                                        titleWidth={150}
                                    >
                                        <DatePicker
                                            value={notifyFromField.value ?? undefined}
                                            onChange={date => {
                                                if (readOnly) {
                                                    return;
                                                }

                                                notifyFromField.onChange(date ?? null);
                                            }}
                                            aria-invalid={isNotifyFromInvalid}
                                            placeholder="Select date"
                                            allowClear
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.notifyFrom]} />
                                    </InfoBlock>
                                </Field>
                            </>
                        )}
                    </FieldGroup>
                </DialogBody>

                <DialogActionFooter>
                    <Button
                        type="submit"
                        isLoading={isPending}
                        className="min-w-[100px]"
                    >
                        Save
                    </Button>
                </DialogActionFooter>
            </fieldset>
        </form>
    );
}

interface PrefillValues {
    email?: string;
    keyType?: ESslKeyType;
    autoRenew?: boolean;
}

interface Props {
    domain: string;
    isPending: boolean;
    prefill?: PrefillValues;
    readOnly?: boolean;
    onSubmit: (values: QuickInstallSslCertFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}
