import { useEffect, useMemo } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { ProjectKeyAuthQueries } from "~/projects/data/queries";
import { KeyAuthQueries } from "~/settings/data/queries";
import { ERegistryAuthKind } from "~/settings/domain";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import {
    AppLink,
    AvailableInAppsWarning,
    Combobox,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import type { RegistryAuthTableScope } from "../registry-auth-table/registry-auth-table.types";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditRegistryAuthFormInput,
    CreateOrEditRegistryAuthFormOutput,
} from "./create-or-edit-registry-auth.form.schema";
import { CreateOrEditRegistryAuthFormSchema, ecrRegionOf } from "./create-or-edit-registry-auth.form.schema";

/** Every key auth the scope can see: a picker offers the list whole. */
const KEY_AUTHS_ALL = { page: 1, size: 1000 };

type KeyAuthOption = {
    id: string;
    name: string;
};

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
    scope = { type: "settings" },
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
            ecrKeyAuth: initialValues?.ecrKeyAuth ?? null,
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
    const { field: ecrKeyAuth } = useController({ name: "ecrKeyAuth", control });
    const {
        field: ecrRoleArn,
        fieldState: { invalid: isEcrRoleArnInvalid },
    } = useController({ name: "ecrRoleArn", control });
    const { field: readonly } = useController({ name: "readonly", control });

    const isEcr = kind.value === ERegistryAuthKind.AwsEcr;
    const ecrRegion = isEcr ? ecrRegionOf(address.value) : "";

    const globalKeyAuthQuery = KeyAuthQueries.useFindManyPaginated(
        { pagination: KEY_AUTHS_ALL },
        { enabled: isEcr && scope.type === "settings" },
    );
    const projectKeyAuthQuery = ProjectKeyAuthQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            pagination: KEY_AUTHS_ALL,
        },
        { enabled: isEcr && scope.type === "project" },
    );
    const keyAuthQuery = scope.type === "project" ? projectKeyAuthQuery : globalKeyAuthQuery;
    const keyAuthOptions = useMemo(
        () =>
            (keyAuthQuery.data?.data ?? []).map(item => ({
                value: { id: item.id, name: item.name } satisfies KeyAuthOption,
                label: item.name,
            })),
        [keyAuthQuery.data?.data],
    );
    const keyAuthManageRoute =
        scope.type === "project"
            ? ROUTE.projects.single.providerConfiguration.keyAuth.$route(scope.projectId)
            : ROUTE.settings.keyAuth.$route;
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
                                        ? "The registry of your AWS account: <account>.dkr.ecr.<region>.amazonaws.com, or the dual-stack <account>.dkr-ecr.<region>.on.aws. Its region is read from it."
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
                                        label="Key Auth"
                                        isRequired
                                        content="The access key ID and secret access key of an IAM user, kept as a Key Auth setting of its own. It needs ecr:GetAuthorizationToken, and the permissions to pull from the repositories - and to push, for built images."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Combobox<KeyAuthOption>
                                            options={keyAuthOptions}
                                            value={ecrKeyAuth.value?.id ?? null}
                                            onChange={(_, option) => {
                                                ecrKeyAuth.onChange(option ?? null);
                                            }}
                                            placeholder="select key auth"
                                            searchable
                                            closeOnSelect
                                            emptyText="No key auths available"
                                            valueKey="id"
                                            loading={keyAuthQuery.isFetching}
                                            onRefresh={() => void keyAuthQuery.refetch()}
                                            isRefreshing={keyAuthQuery.isRefetching}
                                            disabled={isReadOnly}
                                        />
                                        <AppLink.Modules
                                            to={keyAuthManageRoute}
                                            className="text-xs text-link"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            ignorePrevPath
                                        >
                                            Configure Key Auths
                                        </AppLink.Modules>
                                        <FieldError errors={[errors.ecrKeyAuth]} />
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
    /** Where the credential is: the key auths offered are those it can see. */
    scope?: RegistryAuthTableScope;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
