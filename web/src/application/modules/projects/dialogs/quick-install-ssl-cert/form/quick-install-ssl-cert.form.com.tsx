import React, { useEffect, useMemo } from "react";

import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { Field, FieldError, FieldGroup, FieldLabel } from "@components/ui/field";
import { Input } from "@components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { useController, useForm, useWatch } from "react-hook-form";

import { DatePicker, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESslCertType, ESslKeyType } from "@application/shared/enums";

import {
    type QuickInstallSslCertFormInput,
    type QuickInstallSslCertFormOutput,
    QuickInstallSslCertFormSchema,
} from "../schemas";

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

export function QuickInstallSslCertForm({ domain, isPending, prefill, onSubmit, onHasChanges }: Props) {
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
            certType: ESslCertType.LetsEncrypt,
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
            certType: ESslCertType.LetsEncrypt,
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
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const certType = useWatch({ control, name: "certType" });
    const expireAt = useWatch({ control, name: "expireAt" });
    const notifyFrom = useWatch({ control, name: "notifyFrom" });
    const isLetsEncrypt = certType === ESslCertType.LetsEncrypt;

    useEffect(() => {
        if (isLetsEncrypt) {
            setValue("keyType", ESslKeyType.ECP256);
        }
    }, [isLetsEncrypt, setValue]);

    useEffect(() => {
        if (!expireAt || notifyFrom) {
            return;
        }
        setValue("notifyFrom", addDays(expireAt, -30), { shouldDirty: true });
    }, [expireAt, notifyFrom, setValue]);

    const keyTypeOptions = useMemo(() => {
        return (isLetsEncrypt ? LETS_ENCRYPT_KEY_TYPES : CUSTOM_KEY_TYPES).map(value => ({
            value,
            label: formatKeyTypeLabel(value),
        }));
    }, [isLetsEncrypt]);

    const { field: certTypeField } = useController({ name: "certType", control });
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
        field: expireAtField,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const {
        field: notifyFromField,
        fieldState: { invalid: isNotifyFromInvalid },
    } = useController({ name: "notifyFrom", control });

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onSubmit)(event);
            }}
            className="flex flex-col gap-6"
        >
            <FieldGroup>
                <Field>
                    <FieldLabel>Domain</FieldLabel>
                    <Input
                        value={domain}
                        disabled
                    />
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
                        <Tabs
                            value={certType}
                            onValueChange={value => {
                                certTypeField.onChange(value);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger
                                    value={ESslCertType.LetsEncrypt}
                                    className="flex-1"
                                >
                                    Let&apos;s Encrypt
                                </TabsTrigger>
                                <TabsTrigger
                                    value={ESslCertType.Custom}
                                    className="flex-1"
                                >
                                    Custom
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>
                </Field>

                <Field>
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label={isLetsEncrypt ? "Registration E-mail" : "E-mail"}
                                isRequired
                            />
                        }
                        titleWidth={150}
                    >
                        <Input
                            {...email}
                            type="email"
                            aria-invalid={isEmailInvalid}
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
                                keyType.onChange(value);
                            }}
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

                {isLetsEncrypt ? (
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
                                    autoRenew.onChange(value === true);
                                }}
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
                                />
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
                                        expireAtField.onChange(date ?? null);
                                    }}
                                    aria-invalid={isExpireAtInvalid}
                                    placeholder="Select date"
                                    allowClear
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
                                        notifyFromField.onChange(date ?? null);
                                    }}
                                    aria-invalid={isNotifyFromInvalid}
                                    placeholder="Select date"
                                    allowClear
                                />
                                <FieldError errors={[errors.notifyFrom]} />
                            </InfoBlock>
                        </Field>
                    </>
                )}
            </FieldGroup>

            <div className="flex justify-end">
                <Button
                    type="submit"
                    isLoading={isPending}
                >
                    Save
                </Button>
            </div>
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
    onSubmit: (values: QuickInstallSslCertFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}
