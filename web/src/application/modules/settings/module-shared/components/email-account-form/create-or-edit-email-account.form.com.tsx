import { useEffect, useState } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useFieldArray, useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import {
    SETTINGS_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS,
    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
    SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS,
} from "~/settings/module-shared/constants/settings-form-layout.constants";
import { InfoBlock, InputWithAddOn, LabelWithInfo, SelectWithAddon } from "@application/shared/components";
import { EEmailKind } from "@application/shared/enums";
import { FieldListLayout, KeyValueList } from "@application/shared/form";

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

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import {
    type CreateOrEditEmailAccountFormInput,
    type CreateOrEditEmailAccountFormOutput,
    CreateOrEditEmailAccountFormSchema,
    EMAIL_HTTP_METHODS,
} from "./create-or-edit-email-account.form.schema";

const FIELD_MAPPING_NAMES = [
    "fromAddress",
    "fromName",
    "toAddress",
    "toAddresses",
    "subject",
    "content",
    "password",
] as const;

const FIELD_MAPPING_OPTIONS = FIELD_MAPPING_NAMES.map(name => ({
    label: name,
    value: name,
}));

export function CreateOrEditEmailAccountForm({
    isPending,
    onSubmit,
    onOpenTestSendMail,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const form = useForm<CreateOrEditEmailAccountFormInput, unknown, CreateOrEditEmailAccountFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? EEmailKind.SMTP,
            smtpHost: initialValues?.smtpHost ?? "",
            smtpPort: initialValues?.smtpPort ?? 587,
            smtpUsername: initialValues?.smtpUsername ?? "",
            smtpDisplayName: initialValues?.smtpDisplayName ?? "",
            smtpPassword: initialValues?.smtpPassword ?? "",
            httpEndpoint: initialValues?.httpEndpoint ?? "",
            httpMethod: initialValues?.httpMethod ?? "POST",
            httpUsername: initialValues?.httpUsername ?? "",
            httpDisplayName: initialValues?.httpDisplayName ?? "",
            httpPassword: initialValues?.httpPassword ?? "",
            httpContentType: initialValues?.httpContentType ?? "application/json",
            headers: initialValues?.headers ?? [],
            fieldMapping: initialValues?.fieldMapping ?? [],
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditEmailAccountFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = form;

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, onHasChanges, isReadOnly]);

    const kindValue = useWatch({ control, name: "kind" });
    const fieldMapping = useFieldArray({ control, name: "fieldMapping" });
    const [mappingKey, setMappingKey] = useState("");
    const [mappingValue, setMappingValue] = useState("");

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const { field: kind } = useController({ name: "kind", control });
    const {
        field: smtpHost,
        fieldState: { invalid: isSmtpHostInvalid },
    } = useController({ name: "smtpHost", control });
    const {
        field: smtpPort,
        fieldState: { invalid: isSmtpPortInvalid },
    } = useController({ name: "smtpPort", control });
    const {
        field: smtpUsername,
        fieldState: { invalid: isSmtpUsernameInvalid },
    } = useController({ name: "smtpUsername", control });
    const { field: smtpDisplayName } = useController({ name: "smtpDisplayName", control });
    const {
        field: smtpPassword,
        fieldState: { invalid: isSmtpPasswordInvalid },
    } = useController({ name: "smtpPassword", control });
    const {
        field: httpEndpoint,
        fieldState: { invalid: isHttpEndpointInvalid },
    } = useController({ name: "httpEndpoint", control });
    const {
        field: httpMethod,
        fieldState: { invalid: isHttpMethodInvalid },
    } = useController({ name: "httpMethod", control });
    const { field: httpUsername } = useController({ name: "httpUsername", control });
    const { field: httpDisplayName } = useController({ name: "httpDisplayName", control });
    const { field: httpPassword } = useController({ name: "httpPassword", control });
    const { field: httpContentType } = useController({ name: "httpContentType", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditEmailAccountFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditEmailAccountFormOutput) {
        onOpenTestSendMail(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditEmailAccountFormOutput>) {
        console.error(_errors);
    }

    function handleAddFieldMapping() {
        const trimmedKey = mappingKey.trim();
        if (!trimmedKey) {
            return;
        }

        const exists = fieldMapping.fields.some(field => field.key === trimmedKey);
        if (exists) {
            toast.error(`Key "${trimmedKey}" already exists`);
            return;
        }

        fieldMapping.append({ key: trimmedKey, value: mappingValue.trim() });
        setMappingKey("");
        setMappingValue("");
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
                        <FieldGroup>
                            <InfoBlock
                                titleWidth={220}
                                title={<LabelWithInfo label="Name" />}
                            >
                                <Field>
                                    <Input
                                        {...name}
                                        aria-invalid={isNameInvalid}
                                    />
                                    <FieldError errors={[errors.name]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={<LabelWithInfo label="Email Type" />}
                            >
                                <Tabs
                                    value={kind.value}
                                    onValueChange={value => {
                                        kind.onChange(value);
                                    }}
                                >
                                    <TabsList>
                                        <TabsTrigger value={EEmailKind.SMTP}>SMTP</TabsTrigger>
                                        <TabsTrigger value={EEmailKind.HTTP}>HTTP</TabsTrigger>
                                    </TabsList>
                                </Tabs>
                            </InfoBlock>

                            {kindValue === EEmailKind.SMTP && (
                                <>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Host"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                {...smtpHost}
                                                aria-invalid={isSmtpHostInvalid}
                                            />
                                            <FieldError errors={[errors.smtpHost]} />
                                        </Field>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Port"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                type="number"
                                                min={1}
                                                max={65535}
                                                {...smtpPort}
                                                aria-invalid={isSmtpPortInvalid}
                                            />
                                            <FieldError errors={[errors.smtpPort]} />
                                        </Field>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Username"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                {...smtpUsername}
                                                aria-invalid={isSmtpUsernameInvalid}
                                            />
                                            <FieldError errors={[errors.smtpUsername]} />
                                        </Field>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Display Name" />}
                                    >
                                        <Input {...smtpDisplayName} />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Password"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <PasswordInput
                                                {...smtpPassword}
                                                aria-invalid={isSmtpPasswordInvalid}
                                            />
                                            <FieldError errors={[errors.smtpPassword]} />
                                        </Field>
                                    </InfoBlock>
                                </>
                            )}

                            {kindValue === EEmailKind.HTTP && (
                                <>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Endpoint"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                {...httpEndpoint}
                                                aria-invalid={isHttpEndpointInvalid}
                                            />
                                            <FieldError errors={[errors.httpEndpoint]} />
                                        </Field>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Method"
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
                                            <Select
                                                value={httpMethod.value}
                                                onValueChange={value => {
                                                    httpMethod.onChange(value);
                                                }}
                                            >
                                                <SelectTrigger aria-invalid={isHttpMethodInvalid}>
                                                    <SelectValue placeholder="Select method" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {EMAIL_HTTP_METHODS.map(method => (
                                                        <SelectItem
                                                            key={method}
                                                            value={method}
                                                        >
                                                            {method}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            <FieldError errors={[errors.httpMethod]} />
                                        </Field>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Username" />}
                                    >
                                        <Input {...httpUsername} />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Display Name" />}
                                    >
                                        <Input {...httpDisplayName} />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Password" />}
                                    >
                                        <div className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                            <PasswordInput {...httpPassword} />
                                        </div>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Content Type" />}
                                    >
                                        <Input {...httpContentType} />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Headers" />}
                                    >
                                        <KeyValueList<CreateOrEditEmailAccountFormInput>
                                            name="headers"
                                            keyLabel="Name"
                                            valueLabel="Value"
                                            keyPlaceholder="name"
                                            valuePlaceholder="value"
                                            checkDuplicates
                                            className={SETTINGS_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS}
                                        />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Field Mappings" />}
                                    >
                                        <FieldListLayout
                                            inputsClassName={`grid flex-1 grid-cols-2 gap-2 ${SETTINGS_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS}`}
                                            inputRow={
                                                <>
                                                    <SelectWithAddon
                                                        addonLeft="Name"
                                                        value={mappingKey}
                                                        onValueChange={setMappingKey}
                                                        placeholder="select field"
                                                    >
                                                        {FIELD_MAPPING_OPTIONS.map(option => (
                                                            <SelectItem
                                                                key={option.value}
                                                                value={option.value}
                                                            >
                                                                {option.label}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectWithAddon>
                                                    <InputWithAddOn
                                                        addonLeft="Value"
                                                        value={mappingValue}
                                                        onChange={event => {
                                                            setMappingValue(event.target.value);
                                                        }}
                                                        placeholder="value"
                                                    />
                                                </>
                                            }
                                            onAdd={handleAddFieldMapping}
                                            addDisabled={!mappingKey.trim()}
                                            items={fieldMapping.fields.map((field, index) => ({
                                                id: field.id,
                                                content: (
                                                    <div className="grid flex-1 grid-cols-2 gap-2">
                                                        <div className="break-words text-sm">{field.key}</div>
                                                        <div className="break-words text-sm">{field.value}</div>
                                                    </div>
                                                ),
                                                onRemove: () => {
                                                    fieldMapping.remove(index);
                                                },
                                            }))}
                                        />
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
                    </fieldset>
                </div>
                {!isReadOnly && (
                    <div className="pb-6 flex justify-between mt-6">
                        <div className="flex items-center gap-3">
                            <Button
                                type="button"
                                variant="secondary"
                                onClick={() => {
                                    void handleSubmit(onTestValid, onInvalid)();
                                }}
                            >
                                Test Send Email
                            </Button>
                        </div>
                        <div className="flex items-center gap-3">
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
        </FormProvider>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditEmailAccountFormOutput) => void;
    onOpenTestSendMail: (values: CreateOrEditEmailAccountFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditEmailAccountFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
