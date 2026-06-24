import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";
import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditRegistryAuthFormInput,
    CreateOrEditRegistryAuthFormOutput,
} from "./create-or-edit-registry-auth.form.schema";
import { CreateOrEditRegistryAuthFormSchema } from "./create-or-edit-registry-auth.form.schema";

export function CreateOrEditRegistryAuthForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestConnection,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditRegistryAuthFormInput, unknown, CreateOrEditRegistryAuthFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            address: initialValues?.address ?? "",
            username: initialValues?.username ?? "",
            password: initialValues?.password ?? "",
            readonly: initialValues?.readonly ?? false,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditRegistryAuthFormSchema),
        mode: "onSubmit",
    });

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
    const { field: readonly } = useController({ name: "readonly", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
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
                                label="Server Address"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...address}
                                    aria-invalid={isAddressInvalid}
                                />
                                <FieldError errors={[errors.address]} />
                            </Field>
                        </FieldGroup>
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
                <div className="pb-6 flex justify-between mt-6">
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
    );
}

interface Props {
    isPending: boolean;
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditRegistryAuthFormOutput) => void;
    onTestConnection: (values: CreateOrEditRegistryAuthFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditRegistryAuthFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
