import { useEffect, useMemo, useState } from "react";

import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { ProjectNotificationQueries } from "~/projects/data/queries";
import { NotificationQueries } from "~/settings/data/queries";
import type { SslCertTableScope } from "~/settings/module-shared/components";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { Combobox, ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { ESslCertType, ESslKeyType } from "@application/shared/enums";

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
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import type { CreateOrEditSslCertFormInput, CreateOrEditSslCertFormOutput } from "../schemas";
import { CreateOrEditSslCertFormSchema } from "../schemas";

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

type NotificationFieldName = "notification.success" | "notification.failure";

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

function NotificationSelect({ control, disabled, name, scope, title }: NotificationSelectProps) {
    const [searchQuery, setSearchQuery] = useState("");

    const settingsQuery = NotificationQueries.useFindManyPaginated(
        {
            search: searchQuery,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectNotificationQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            search: searchQuery,
        },
        {
            enabled: scope.type === "project",
        },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: notifications } = DEFAULT_PAGINATED_DATA, isFetching, refetch, isRefetching } = query;

    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name });

    const comboboxOptions = useMemo(() => {
        return notifications.map(notification => ({
            value: { id: notification.id, name: notification.name },
            label: notification.name,
        }));
    }, [notifications]);

    return (
        <InfoBlock
            title={title}
            titleWidth={220}
        >
            <Field>
                <Combobox
                    options={comboboxOptions}
                    value={field.value?.id ?? null}
                    onChange={(_, option) => {
                        field.onChange(option ?? undefined);
                    }}
                    onSearch={setSearchQuery}
                    placeholder="None"
                    searchable
                    closeOnSelect
                    emptyText="No notifications available"
                    valueKey="id"
                    aria-invalid={invalid}
                    loading={isFetching}
                    disabled={disabled}
                    onRefresh={() => void refetch()}
                    isRefreshing={isRefetching}
                />
                <FieldError errors={[error]} />
            </Field>
        </InfoBlock>
    );
}

export function CreateOrEditSslCertForm({
    isPending,
    onSubmit,
    onHasChanges,
    initialValues,
    scope,
    showAvailableInProjects,
    readOnlyInherited = false,
    onClose,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
        reset,
        setValue,
    } = useForm<CreateOrEditSslCertFormInput, unknown, CreateOrEditSslCertFormOutput>({
        defaultValues: {
            domain: initialValues?.domain ?? "",
            certType: initialValues?.certType ?? ESslCertType.LetsEncrypt,
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

    useEffect(() => {
        reset({
            domain: initialValues?.domain ?? "",
            certType: initialValues?.certType ?? ESslCertType.LetsEncrypt,
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
        reset,
    ]);

    useEffect(() => {
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty, onHasChanges, readOnlyInherited]);

    const certType = useWatch({ control, name: "certType" });
    const expireAt = useWatch({ control, name: "expireAt" });
    const notifyFrom = useWatch({ control, name: "notifyFrom" });
    const successUseDefault = useWatch({ control, name: "notification.successUseDefault" });
    const success = useWatch({ control, name: "notification.success" });
    const failureUseDefault = useWatch({ control, name: "notification.failureUseDefault" });
    const failure = useWatch({ control, name: "notification.failure" });
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

    useEffect(() => {
        if (successUseDefault && success) {
            setValue("notification.success", undefined);
        }
    }, [setValue, success, successUseDefault]);

    useEffect(() => {
        if (failureUseDefault && failure) {
            setValue("notification.failure", undefined);
        }
    }, [failure, failureUseDefault, setValue]);

    const keyTypeOptions = useMemo(() => {
        return (isLetsEncrypt ? LETS_ENCRYPT_KEY_TYPES : CUSTOM_KEY_TYPES).map(value => ({
            value,
            label: formatKeyTypeLabel(value),
        }));
    }, [isLetsEncrypt]);

    const {
        field: domain,
        fieldState: { invalid: isDomainInvalid },
    } = useController({ name: "domain", control });
    const { field: certTypeField } = useController({ name: "certType", control });
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
    const { field: successUseDefaultField } = useController({ name: "notification.successUseDefault", control });
    const { field: failureUseDefaultField } = useController({ name: "notification.failureUseDefault", control });

    function onValid(values: CreateOrEditSslCertFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditSslCertFormOutput>) {
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
            {readOnlyInherited && <InheritedSettingReadonlyNotice />}
            <fieldset
                disabled={readOnlyInherited}
                className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
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
                        <Tabs
                            value={certType}
                            onValueChange={value => {
                                certTypeField.onChange(value);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger value={ESslCertType.LetsEncrypt}>Let&apos;s Encrypt</TabsTrigger>
                                <TabsTrigger value={ESslCertType.Custom}>Custom</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label={isLetsEncrypt ? "Registration E-mail" : "E-mail"}
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

                    {isLetsEncrypt ? (
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
                                <Field>
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
                                <Field>
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
                    <div className="flex flex-col gap-6">
                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="On Success Use Default"
                                    content="Use the default notification settings on success"
                                />
                            }
                        >
                            <Checkbox
                                checked={successUseDefaultField.value}
                                onCheckedChange={checked => {
                                    successUseDefaultField.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>

                        <NotificationSelect
                            control={control}
                            disabled={successUseDefault}
                            name="notification.success"
                            scope={scope}
                            title="On Success"
                        />

                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="On Failure Use Default"
                                    content="Use the default notification settings on failure"
                                />
                            }
                        >
                            <Checkbox
                                checked={failureUseDefaultField.value}
                                onCheckedChange={checked => {
                                    failureUseDefaultField.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>

                        <NotificationSelect
                            control={control}
                            disabled={failureUseDefault}
                            name="notification.failure"
                            scope={scope}
                            title="On Failure"
                        />
                    </div>
                </ContentBlock>

                {!readOnlyInherited && (
                    <Field>
                        <div className="flex justify-end">
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
    );
}

interface NotificationSelectProps {
    control: ReturnType<
        typeof useForm<CreateOrEditSslCertFormInput, unknown, CreateOrEditSslCertFormOutput>
    >["control"];
    disabled: boolean;
    name: NotificationFieldName;
    scope: SslCertTableScope;
    title: string;
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditSslCertFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditSslCertFormInput>;
    scope: SslCertTableScope;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}
