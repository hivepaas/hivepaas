import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditBasicAuthFormInput, CreateOrEditBasicAuthFormOutput } from "../schemas";
import { CreateOrEditBasicAuthFormSchema } from "../schemas";

export function CreateOrEditBasicAuthForm({
    isPending,
    onSubmit,
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
    } = useForm<CreateOrEditBasicAuthFormInput, unknown, CreateOrEditBasicAuthFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            username: initialValues?.username ?? "",
            password: "",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditBasicAuthFormSchema),
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
        field: username,
        fieldState: { invalid: isUsernameInvalid },
    } = useController({ name: "username", control });
    const {
        field: password,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({ name: "password", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditBasicAuthFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditBasicAuthFormOutput>) {
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

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditBasicAuthFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditBasicAuthFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}
