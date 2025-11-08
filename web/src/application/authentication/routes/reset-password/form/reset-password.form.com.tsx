import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { z } from "zod";

import { PasswordStrengthMeter } from "@application/shared/components";

import { BackToSignIn } from "@application/authentication/components";

const Schema = z
    .object({
        newPassword: z.string().trim().min(1, "Password is required"),
        confirmPassword: z.string().trim().min(1, "Confirm Password is required"),
        isStrongPassword: z.boolean(),
    })
    .superRefine((arg, ctx) => {
        if (arg.newPassword !== arg.confirmPassword) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Passwords do not match",
                path: ["confirmPassword"],
            });
        }

        if (arg.newPassword.length > 0 && !arg.isStrongPassword) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Password is not strong enough",
                path: ["isStrongPassword"],
            });
        }
    });

type SchemaInput = z.input<typeof Schema>;
type SchemaOutput = z.output<typeof Schema>;

export function ResetPasswordForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isValid },
    } = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            newPassword: "",
            confirmPassword: "",
            isStrongPassword: false,
        },
        resolver: zodResolver(Schema),
        mode: "onSubmit",
    });

    const {
        field: newPassword,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({
        name: "newPassword",
        control,
    });

    const {
        field: confirmPassword,
        fieldState: { invalid: isConfirmPasswordInvalid },
    } = useController({
        name: "confirmPassword",
        control,
    });

    const {
        field: isStrongPassword,
        fieldState: { invalid: isNotStrongEnough },
    } = useController({
        name: "isStrongPassword",
        control,
    });

    function onValid(values: SchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(_: FieldErrors<SchemaOutput>) {
        //
    }

    return (
        <div className="flex flex-col gap-6">
            <Card>
                <CardHeader>
                    <CardTitle>Reset password</CardTitle>
                    <CardDescription>Enter your new password</CardDescription>
                </CardHeader>
                <CardContent>
                    <form
                        onSubmit={event => {
                            event.preventDefault();

                            void handleSubmit(onValid, onInvalid)(event);
                        }}
                    >
                        <FieldGroup>
                            <Field>
                                <FieldLabel htmlFor="password">New password</FieldLabel>
                                <PasswordInput
                                    id="newPassword"
                                    value={newPassword.value}
                                    onChange={newPassword.onChange}
                                    aria-invalid={isPasswordInvalid || isNotStrongEnough}
                                    required
                                />
                                <FieldError errors={[errors.newPassword, errors.isStrongPassword]} />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="confirmPassword">Confirm password</FieldLabel>
                                <PasswordInput
                                    id="confirmPassword"
                                    value={confirmPassword.value}
                                    onChange={confirmPassword.onChange}
                                    aria-invalid={isConfirmPasswordInvalid || isNotStrongEnough}
                                    required
                                />
                                <FieldError errors={[errors.confirmPassword, errors.isStrongPassword]} />
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
                                    disabled={!isValid}
                                >
                                    Reset Password
                                </Button>
                                <BackToSignIn />
                            </Field>
                        </FieldGroup>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: SchemaOutput) => Promise<void> | void;
}
