import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { ERegistryAuthKind } from "~/settings/domain";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { AvailableInAppsWarning, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditRegistryAuthFormInput,
    CreateOrEditRegistryAuthFormOutput,
} from "./create-or-edit-registry-auth.form.schema";
import { CreateOrEditRegistryAuthFormSchema, ecrRegionOf } from "./create-or-edit-registry-auth.form.schema";

export function CreateOrEditRegistryAuthForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestConnection,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    ecrTokenExpiresAt,
    showAvailableInProjects = true,
    isProjectScope = false,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;
    const isInheritableDisabled = isReadOnly;
    const inheritableLabel = isProjectScope ? "Available in Apps" : "Available in Projects";

    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditRegistryAuthFormInput, unknown, CreateOrEditRegistryAuthFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? ERegistryAuthKind.Basic,
            address: initialValues?.address ?? "",
            username: initialValues?.username ?? "",
            password: initialValues?.password ?? "",
            ecrAccessKeyId: initialValues?.ecrAccessKeyId ?? "",
            ecrSecretAccessKey: initialValues?.ecrSecretAccessKey ?? "",
            ecrRoleArn: initialValues?.ecrRoleArn ?? "",
            readonly: initialValues?.readonly ?? false,
            inheritable: initialValues?.inheritable ?? (isProjectScope ? true : false),
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditRegistryAuthFormSchema),
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

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: address,
        fieldState: { invalid: isAddressInvalid },
    } = useController({ name: "address", control });
    const {
        field: username,
        fieldState: { invalid: isUsernameInvalid },
    } = useController({ name: "username", control });
    const {
        field: password,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({ name: "password", control });
    const { field: kind } = useController({ name: "kind", control });
    const {
        field: ecrAccessKeyId,
        fieldState: { invalid: isEcrAccessKeyIdInvalid },
    } = useController({ name: "ecrAccessKeyId", control });
    const {
        field: ecrSecretAccessKey,
        fieldState: { invalid: isEcrSecretAccessKeyInvalid },
    } = useController({ name: "ecrSecretAccessKey", control });
    const {
        field: ecrRoleArn,
        fieldState: { invalid: isEcrRoleArnInvalid },
    } = useController({ name: "ecrRoleArn", control });
    const { field: readonly } = useController({ name: "readonly", control });

    const isEcr = kind.value === ERegistryAuthKind.AwsEcr;
    const ecrRegion = isEcr ? ecrRegionOf(address.value) : "";
    const { field: inheritable } = useController({ name: "inheritable", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditRegistryAuthFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditRegistryAuthFormOutput) {
        onTestConnection(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditRegistryAuthFormOutput>) {
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
                                content="Amazon ECR has no password that lasts: HivePaaS gets a token from AWS keys, which expires after 12 hours, and renews it in the services that pull with it."
                            />
                        }
                    >
                        <Tabs
                            value={kind.value}
                            onValueChange={kind.onChange}
                        >
                            <TabsList>
                                <TabsTrigger value={ERegistryAuthKind.Basic}>Username and password</TabsTrigger>
                                <TabsTrigger value={ERegistryAuthKind.AwsEcr}>Amazon ECR</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Server Address"
                                isRequired
                                content={
                                    isEcr
                                        ? "The registry of your AWS account: <account>.dkr.ecr.<region>.amazonaws.com. Its region is read from it."
                                        : undefined
                                }
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...address}
                                    placeholder={isEcr ? "123456789012.dkr.ecr.eu-west-1.amazonaws.com" : undefined}
                                    aria-invalid={isAddressInvalid}
                                />
                                {ecrRegion && <p className="text-xs text-muted-foreground">Region: {ecrRegion}</p>}
                                <FieldError errors={[errors.address]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    {isEcr && (
                        <>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Access Key ID"
                                        isRequired
                                        content="An IAM user's access key. It needs ecr:GetAuthorizationToken, and the permissions to pull from the repositories - and to push, for built images."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...ecrAccessKeyId}
                                            autoComplete="off"
                                            aria-invalid={isEcrAccessKeyIdInvalid}
                                        />
                                        <FieldError errors={[errors.ecrAccessKeyId]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Secret Access Key"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <PasswordInput
                                            value={ecrSecretAccessKey.value}
                                            onChange={ecrSecretAccessKey.onChange}
                                            aria-invalid={isEcrSecretAccessKeyInvalid}
                                        />
                                        <FieldError errors={[errors.ecrSecretAccessKey]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Role ARN"
                                        content="Optional: an IAM role assumed with the keys above, which then carries the ECR permissions."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...ecrRoleArn}
                                            placeholder="arn:aws:iam::123456789012:role/hivepaas-pull"
                                            aria-invalid={isEcrRoleArnInvalid}
                                        />
                                        <FieldError errors={[errors.ecrRoleArn]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            {ecrTokenExpiresAt !== undefined && (
                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="Token Expires At"
                                            content="The token got from the keys, kept until it is too old to hand to Swarm. The Registry Auth Renewal settings renew it."
                                        />
                                    }
                                >
                                    <span className="text-sm">
                                        {ecrTokenExpiresAt
                                            ? format(ecrTokenExpiresAt, "yyyy-MM-dd HH:mm:ss")
                                            : "No token yet: one is got at the first deploy, build or renewal"}
                                    </span>
                                </InfoBlock>
                            )}
                        </>
                    )}

                    {!isEcr && (
                        <>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Username"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...username}
                                            aria-invalid={isUsernameInvalid}
                                        />
                                        <FieldError errors={[errors.username]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Password"
                                        isRequired
                                        content="For Google Artifact Registry: username _json_key_base64, and the service account's JSON key, in base64, as the password."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <PasswordInput
                                            value={password.value}
                                            onChange={password.onChange}
                                            aria-invalid={isPasswordInvalid}
                                        />
                                        <FieldError errors={[errors.password]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </>
                    )}

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Read Only" />}
                    >
                        <Checkbox
                            checked={readonly.value}
                            onCheckedChange={checked => {
                                readonly.onChange(Boolean(checked));
                            }}
                        />
                    </InfoBlock>

                    {showAvailableInProjects && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label={inheritableLabel} />}
                        >
                            <div className="flex items-center gap-3">
                                <Checkbox
                                    disabled={isInheritableDisabled}
                                    checked={inheritable.value}
                                    onCheckedChange={checked => {
                                        inheritable.onChange(Boolean(checked));
                                    }}
                                />
                                {isProjectScope && !inheritable.value ? <AvailableInAppsWarning /> : null}
                            </div>
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
                <FormActionBar contentClassName="justify-between">
                    <div className="flex items-center gap-3">
                        <Button
                            type="button"
                            variant="secondary"
                            isLoading={isTesting}
                            onClick={() => {
                                void handleSubmit(onTestValid, onInvalid)();
                            }}
                        >
                            Test Connection
                        </Button>
                        {testStatus === "succeeded" && <span className="text-sm text-green-600">Succeeded</span>}
                        {testStatus === "failed" && <span className="text-sm text-destructive">Failed</span>}
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
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditRegistryAuthFormOutput) => void;
    onTestConnection: (values: CreateOrEditRegistryAuthFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditRegistryAuthFormInput>;
    /** An Amazon ECR credential's token expiry, on edit: null before one is got. */
    ecrTokenExpiresAt?: Date | null;
    showAvailableInProjects?: boolean;
    isProjectScope?: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
