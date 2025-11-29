import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";

import { PasswordStrengthMeter } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";

import {
    AccountPasswordFormSchema,
    type AccountPasswordFormSchemaInput,
    type AccountPasswordFormSchemaOutput,
} from "../schemas";

export function ChangePasswordForm({ isPending, onSubmit, onHasChanges }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<AccountPasswordFormSchemaInput, unknown, AccountPasswordFormSchemaOutput>({
        defaultValues: {
            currentPassword: "",
            newPassword: "",
            confirmNewPassword: "",
            isStrongPassword: false,
        },
        resolver: zodResolver(AccountPasswordFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty]);

    const {
        field: currentPassword,
        fieldState: { invalid: isCurrentPasswordInvalid },
    } = useController({
        name: "currentPassword",
        control,
    });

    const {
        field: newPassword,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({
        name: "newPassword",
        control,
    });

    const {
        field: confirmNewPassword,
        fieldState: { invalid: isConfirmPasswordInvalid },
    } = useController({
        name: "confirmNewPassword",
        control,
    });

    const {
        field: isStrongPassword,
        fieldState: { invalid: isNotStrongEnough },
    } = useController({
        name: "isStrongPassword",
        control,
    });

    function onValid(values: AccountPasswordFormSchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(fieldErrors: FieldErrors<AccountPasswordFormSchemaOutput>) {
        console.error(fieldErrors);
    }

    return (
        <div className="flex flex-col gap-4">
            <form
                onSubmit={event => {
                    event.preventDefault();

                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="currentPassword">Current password</FieldLabel>
                        <PasswordInput
                            id="currentPassword"
                            value={currentPassword.value}
                            onChange={currentPassword.onChange}
                            aria-invalid={isCurrentPasswordInvalid}
                        />
                        <FieldError errors={[errors.currentPassword]} />
                    </Field>
                    <Field>
                        <FieldLabel htmlFor="newPassword">New password</FieldLabel>
                        <PasswordInput
                            id="newPassword"
                            value={newPassword.value}
                            onChange={newPassword.onChange}
                            aria-invalid={isPasswordInvalid || isNotStrongEnough}
                        />
                        <FieldError errors={[errors.newPassword, errors.isStrongPassword]} />
                    </Field>
                    <Field>
                        <FieldLabel htmlFor="confirmNewPassword">Confirm new password</FieldLabel>
                        <PasswordInput
                            id="confirmNewPassword"
                            value={confirmNewPassword.value}
                            onChange={confirmNewPassword.onChange}
                            aria-invalid={isConfirmPasswordInvalid || isNotStrongEnough}
                        />
                        <FieldError errors={[errors.confirmNewPassword, errors.isStrongPassword]} />
                    </Field>
                    <Field>
                        <PasswordStrengthMeter
                            password={newPassword.value}
                            onStrengthChange={strength => {
                                isStrongPassword.onChange(strength === "max");
                            }}
                        />
                    </Field>
                    <Field>
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Change
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: AccountPasswordFormSchemaOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}
