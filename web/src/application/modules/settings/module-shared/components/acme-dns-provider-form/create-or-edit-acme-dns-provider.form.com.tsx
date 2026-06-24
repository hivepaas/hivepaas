import { useEffect, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { ACME_DNS_PROVIDER_OPTIONS } from "~/settings/module-shared/constants/acme-dns-provider.constants";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EAcmeDnsProviderKind } from "@application/shared/enums";

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

import { AcmeDnsProviderFields, TestAcmeDnsProviderAccessDialog } from "./building-blocks";
import {
    type CreateOrEditAcmeDnsProviderFormInput,
    type CreateOrEditAcmeDnsProviderFormOutput,
    CreateOrEditAcmeDnsProviderFormSchema,
} from "./create-or-edit-acme-dns-provider.form.schema";

export const DEFAULT_ACME_DNS_PROVIDER_FORM_VALUES: CreateOrEditAcmeDnsProviderFormInput = {
    name: "",
    kind: EAcmeDnsProviderKind.AcmeDNS,
    acmeDnsApiBase: "",
    acmeDnsAllowList: "",
    acmeDnsStoragePath: "",
    acmeDnsStorageBaseUrl: "",
    azureClientId: "",
    azureClientSecret: "",
    azureSubscriptionId: "",
    azureTenantId: "",
    azureResourceGroupName: "",
    baiduCloudAccessKey: "",
    baiduCloudSecretKey: "",
    cloudflareAuthToken: "",
    digitalOceanAuthToken: "",
    gCloudServiceAccount: "",
    gCloudProjectId: "",
    goDaddyApiKey: "",
    goDaddyApiSecret: "",
    hetznerApiToken: "",
    huaweiCloudAccessKey: "",
    huaweiCloudSecretKey: "",
    huaweiCloudRegion: "",
    namecheapApiUser: "",
    namecheapApiKey: "",
    rfc2136Nameserver: "",
    rfc2136TsigKeyName: "",
    rfc2136TsigSecret: "",
    rfc2136TsigAlgorithm: "",
    route53AccessKeyId: "",
    route53SecretAccessKey: "",
    route53HostedZoneId: "",
    route53Region: "",
    tencentCloudSecretId: "",
    tencentCloudSecretKey: "",
    tencentCloudRegion: "",
    availableInProjects: false,
    default: false,
};

export function CreateOrEditAcmeDnsProviderForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestAccess,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
    showTestAccess,
    isEdit,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;
    const [testDialogOpen, setTestDialogOpen] = useState(false);
    const [testValues, setTestValues] = useState<CreateOrEditAcmeDnsProviderFormOutput | null>(null);

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditAcmeDnsProviderFormInput, unknown, CreateOrEditAcmeDnsProviderFormOutput>({
        defaultValues: {
            ...DEFAULT_ACME_DNS_PROVIDER_FORM_VALUES,
            ...initialValues,
        },
        resolver: zodResolver(CreateOrEditAcmeDnsProviderFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, isReadOnly, onHasChanges]);

    const kind = useWatch({ control, name: "kind" });
    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kindField,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditAcmeDnsProviderFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditAcmeDnsProviderFormOutput>) {
        console.error(_errors);
    }

    function handleOpenTestAccessDialog() {
        void handleSubmit(values => {
            setTestValues(values);
            setTestDialogOpen(true);
        }, onInvalid)();
    }

    function handleTestAccess(testDomain: string) {
        if (!testValues) {
            return;
        }

        onTestAccess?.(testValues, testDomain);
    }

    return (
        <>
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
                                            <SelectValue placeholder="select provider" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {ACME_DNS_PROVIDER_OPTIONS.map(option => (
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

                        <AcmeDnsProviderFields
                            control={control}
                            errors={errors}
                            kind={kind}
                        />

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
                    <div className="shrink-0 px-0 mt-6 pb-6">
                        <div className="flex flex-1 items-center justify-between gap-4">
                            <div className="flex items-center gap-3">
                                {showTestAccess && (
                                    <Button
                                        type="button"
                                        variant="secondary"
                                        isLoading={isTesting}
                                        onClick={handleOpenTestAccessDialog}
                                    >
                                        Test Access
                                    </Button>
                                )}
                                {testStatus === "success" && <span className="text-sm text-green-500">Succeeded</span>}
                                {testStatus === "error" && <span className="text-sm text-red-500">Failed</span>}
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
            {showTestAccess && (
                <TestAcmeDnsProviderAccessDialog
                    open={testDialogOpen}
                    onOpenChange={setTestDialogOpen}
                    isPending={isTesting}
                    onSubmit={handleTestAccess}
                />
            )}
        </>
    );
}

interface Props {
    isPending: boolean;
    isTesting: boolean;
    testStatus: "idle" | "success" | "error";
    onSubmit: (values: CreateOrEditAcmeDnsProviderFormOutput) => void;
    onTestAccess?: (values: CreateOrEditAcmeDnsProviderFormOutput, testDomain: string) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditAcmeDnsProviderFormInput>;
    showAvailableInProjects: boolean;
    showTestAccess: boolean;
    isEdit: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
