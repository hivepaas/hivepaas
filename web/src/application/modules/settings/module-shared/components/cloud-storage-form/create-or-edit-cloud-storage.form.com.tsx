import { useEffect, useMemo } from "react";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { ProjectKeyAuthQueries } from "~/projects/data/queries";
import { KeyAuthQueries } from "~/settings/data/queries";
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
import { ECloudStorageKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CloudStorageTableScope } from "../cloud-storage-table/cloud-storage-table.types";
import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditCloudStorageFormInput,
    CreateOrEditCloudStorageFormOutput,
} from "./create-or-edit-cloud-storage.form.schema";
import { CreateOrEditCloudStorageFormSchema } from "./create-or-edit-cloud-storage.form.schema";

const providerOptions = Object.values(ECloudStorageKind);

const PROVIDER_LABELS: Record<ECloudStorageKind, string> = {
    [ECloudStorageKind.AWSS3]: "S3 (S3 Compatible)",
};

/** Every key auth the scope can see: a picker offers the list whole. */
const KEY_AUTHS_ALL = { page: 1, size: 1000 };

type KeyAuthOption = {
    id: string;
    name: string;
};

export function CreateOrEditCloudStorageForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestConnection,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    showAvailableInProjects = true,
    scope,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isProjectScope = scope.type === "project";
    const isReadOnly = readOnlyInherited || readOnly;
    const isInheritableDisabled = isReadOnly;
    const inheritableLabel = isProjectScope ? "Available in Apps" : "Available in Projects";

    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditCloudStorageFormInput, unknown, CreateOrEditCloudStorageFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? ECloudStorageKind.AWSS3,
            keyAuth: initialValues?.keyAuth ?? null,
            region: initialValues?.region ?? "",
            bucket: initialValues?.bucket ?? "",
            endpoint: initialValues?.endpoint ?? "",
            inheritable: initialValues?.inheritable ?? (isProjectScope ? true : false),
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditCloudStorageFormSchema),
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
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const { field: keyAuth } = useController({ name: "keyAuth", control });

    const globalKeyAuthQuery = KeyAuthQueries.useFindManyPaginated(
        { pagination: KEY_AUTHS_ALL },
        { enabled: scope.type === "settings" },
    );
    const projectKeyAuthQuery = ProjectKeyAuthQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            pagination: KEY_AUTHS_ALL,
        },
        { enabled: scope.type === "project" },
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
    const {
        field: region,
        fieldState: { invalid: isRegionInvalid },
    } = useController({ name: "region", control });
    const {
        field: bucket,
        fieldState: { invalid: isBucketInvalid },
    } = useController({ name: "bucket", control });
    const {
        field: endpoint,
        fieldState: { invalid: isEndpointInvalid },
    } = useController({ name: "endpoint", control });
    const { field: inheritable } = useController({ name: "inheritable", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditCloudStorageFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditCloudStorageFormOutput) {
        onTestConnection(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditCloudStorageFormOutput>) {
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
                                >
                                    <SelectTrigger aria-invalid={isKindInvalid}>
                                        <SelectValue placeholder="Select provider" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {providerOptions.map(option => (
                                            <SelectItem
                                                key={option}
                                                value={option}
                                            >
                                                {PROVIDER_LABELS[option]}
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
                                label="Key Auth"
                                content="The key ID and secret key the bucket is reached with, kept as a Key Auth setting of its own."
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Combobox<KeyAuthOption>
                                    options={keyAuthOptions}
                                    value={keyAuth.value?.id ?? null}
                                    onChange={(_, option) => {
                                        keyAuth.onChange(option ?? null);
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
                                <FieldError errors={[errors.keyAuth]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Region"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...region}
                                    aria-invalid={isRegionInvalid}
                                />
                                <FieldError errors={[errors.region]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Default Bucket"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...bucket}
                                    aria-invalid={isBucketInvalid}
                                />
                                <FieldError errors={[errors.bucket]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Endpoint" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...endpoint}
                                    aria-invalid={isEndpointInvalid}
                                />
                                <FieldError errors={[errors.endpoint]} />
                            </Field>
                        </FieldGroup>
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
                            Test Access
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
    onSubmit: (values: CreateOrEditCloudStorageFormOutput) => void;
    onTestConnection: (values: CreateOrEditCloudStorageFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditCloudStorageFormInput>;
    showAvailableInProjects?: boolean;
    scope: CloudStorageTableScope;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
