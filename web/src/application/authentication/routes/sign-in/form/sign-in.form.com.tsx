import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { SignInSchema, type SignInSchemaInput, type SignInSchemaOutput } from "../schemas";

export function SignInForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SignInSchemaInput, unknown, SignInSchemaOutput>({
        defaultValues: {
            email: "test@example.com",
            password: "123456",
            rememberMe: false,
        },
        resolver: zodResolver(SignInSchema),
        mode: "onSubmit",
    });

    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({
        name: "email",
        control,
    });

    const {
        field: password,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({
        name: "password",
        control,
    });

    // const { field: rememberMe } = useController({
    //     name: "rememberMe",
    //     control,
    // });

    function onValid(values: SignInSchemaOutput) {
        console.log(values);
        onSubmit(values);
    }

    function onInvalid(errors: FieldErrors<SignInSchemaOutput>) {
        console.log(errors);
    }
    return (
        <div className="flex flex-col gap-6">
            <Card>
                <CardHeader>
                    <CardTitle>Login to your account</CardTitle>
                    <CardDescription>Enter your email below to login to your account</CardDescription>
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
                                <FieldLabel htmlFor="email">Email</FieldLabel>
                                <Input
                                    id="email"
                                    value={email.value}
                                    onChange={email.onChange}
                                    type="email"
                                    placeholder="m@example.com"
                                    aria-invalid={isEmailInvalid}
                                    required
                                />
                            </Field>
                            {isEmailInvalid && <FieldError errors={[errors.email]} />}
                            <Field>
                                <div className="flex items-center">
                                    <FieldLabel htmlFor="password">Password</FieldLabel>
                                    <a
                                        href="#"
                                        className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                                    >
                                        Forgot your password?
                                    </a>
                                </div>
                                <Input
                                    id="password"
                                    value={password.value}
                                    onChange={password.onChange}
                                    type="password"
                                    aria-invalid={isPasswordInvalid}
                                    required
                                />
                            </Field>
                            {isPasswordInvalid && <FieldError errors={[errors.password]} />}
                            <Field>
                                <Button
                                    type="submit"
                                    disabled={isPending}
                                >
                                    Login
                                </Button>
                                <Button
                                    variant="outline"
                                    type="button"
                                >
                                    Login with Google
                                </Button>
                                <FieldDescription className="text-center">
                                    Don&apos;t have an account? <a href="#">Sign up</a>
                                </FieldDescription>
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
    onSubmit: (values: SignInSchemaOutput) => Promise<void> | void;
}
