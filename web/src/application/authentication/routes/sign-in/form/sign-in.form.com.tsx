import { GithubIcon, GitlabIcon, GoogleIcon } from "@/assets/icons";
import { Checkbox } from "@components/ui/checkbox";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { type LoginOption } from "@application/authentication/domain";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { SignInSchema, type SignInSchemaInput, type SignInSchemaOutput } from "../schemas";

export function SignInForm({ loginOptions, isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SignInSchemaInput, unknown, SignInSchemaOutput>({
        defaultValues: {
            email: "tiendc@gmail.com",
            password: "abc123",
            isTrustDevice: true,
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

    const { field: isTrustDevice } = useController({
        name: "isTrustDevice",
        control,
    });

    function onValid(values: SignInSchemaOutput) {
        onSubmit(values);
    }

    function onInvalid(errors: FieldErrors<SignInSchemaOutput>) {
        console.log(errors);
    }

    function handleLoginWithProvider(provider: LoginOption) {
        window.location.href = provider.authURL;
    }
    return (
        <div className="flex flex-col gap-6">
            <Card>
                <CardHeader>
                    <CardTitle>Login to your account</CardTitle>
                    <CardDescription>Enter your account details below to login</CardDescription>
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
                                <FieldLabel htmlFor="email">Username or Email</FieldLabel>
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
                                <div className="flex items-center justify-between">
                                    <FieldLabel htmlFor="password">Password</FieldLabel>
                                    <AppLink
                                        className="text-sm text-primary hover:text-gray-700"
                                        to={ROUTE.auth.forgotPassword.$route}
                                    >
                                        Forgot Password?
                                    </AppLink>
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
                                <div className="flex items-center gap-3">
                                    <Checkbox
                                        id="isTrustDevice"
                                        checked={isTrustDevice.value}
                                        onCheckedChange={isTrustDevice.onChange}
                                    />
                                    <FieldLabel htmlFor="isTrustDevice">Trust this device</FieldLabel>
                                </div>
                            </Field>
                            <Field>
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                >
                                    Login
                                </Button>

                                {loginOptions.map(option => (
                                    <Button
                                        key={option.type}
                                        variant="outline"
                                        type="button"
                                        onClick={() => handleLoginWithProvider(option)}
                                    >
                                        {option.type.toLowerCase().includes("github") && <GithubIcon />}
                                        {option.type.toLowerCase().includes("gitlab") && <GitlabIcon />}
                                        {option.type.toLowerCase().includes("google") && <GoogleIcon />}
                                        Login with {option.name}
                                    </Button>
                                ))}
                            </Field>
                        </FieldGroup>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

interface Props {
    loginOptions: LoginOption[];
    isPending: boolean;
    onSubmit: (values: SignInSchemaOutput) => Promise<void> | void;
}
