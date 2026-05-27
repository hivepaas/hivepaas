import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditRegistryAuthFormInput, CreateOrEditRegistryAuthFormOutput } from "../schemas";
import { CreateOrEditRegistryAuthFormSchema } from "../schemas";

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
    onClose,
}: Props) {
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
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty, onHasChanges, readOnlyInherited]);

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
        if (readOnlyInherited) {
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
            className="flex flex-col gap-6"
        >
            {readOnlyInherited && <InheritedSettingReadonlyNotice />}
            <fieldset
                disabled={readOnlyInherited}
                className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
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

                {!readOnlyInherited && (
                    <Field>
                        <div className="flex items-center justify-between gap-4">
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
                                {testStatus === "succeeded" && (
                                    <span className="text-sm text-green-600">Succeeded</span>
                                )}
                                {testStatus === "failed" && <span className="text-sm text-destructive">Failed</span>}
                            </div>
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
    onClose?: () => void;
}
