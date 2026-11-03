import { GiteaIcon, GithubIcon, GitlabIcon, GoogleIcon } from "@/assets/icons";
import { Checkbox } from "@components/ui/checkbox";
import { zodResolver } from "@hookform/resolvers/zod";
import { DeviceInfo } from "@infrastructure/device";
import { TriangleAlert } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { useSearchParams } from "react-router";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { type LoginOption } from "@application/authentication/domain";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { SignInSchema, type SignInSchemaInput, type SignInSchemaOutput } from "../schemas";

import { ssoErrorMessage } from "./sso-error";

const DEV_SIGN_IN = { email: "tiendc@gmail.com", password: "abc123" };

export function SignInForm({ loginOptions, isPending, onSubmit }: Props) {
    const [params] = useSearchParams();
    const ssoError = params.get("ssoError");

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SignInSchemaInput, unknown, SignInSchemaOutput>({
        defaultValues: {
            // The seeded account is filled in for development only: a built
            // dashboard must not hand every visitor a login.
            ...(import.meta.env.MODE === "development" ? DEV_SIGN_IN : { email: "", password: "" }),
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
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<SignInSchemaOutput>) {
        console.log(_errors);
    }

    async function handleLoginWithProvider(provider: LoginOption) {
        const url = new URL(provider.authURL, window.location.origin);
        // A trusted device is not asked for the passcode again for a while, after
        // a provider's sign-in as after a password.
        const deviceId = isTrustDevice.value ? await new DeviceInfo().getFingerprint() : undefined;
        if (deviceId) {
            url.searchParams.set("trustedDeviceId", deviceId);
        }
        window.location.href = url.toString();
    }
    return (
        <div className="flex flex-col gap-6">
            <Card>
                <CardContent>
                    <form
                        onSubmit={event => {
                            event.preventDefault();

                            void handleSubmit(onValid, onInvalid)(event);
                        }}
                    >
                        <FieldGroup>
                            {ssoError && (
                                <div
                                    role="alert"
                                    className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive"
                                >
                                    <TriangleAlert className="mt-0.5 size-4 shrink-0" />
                                    <span>{ssoErrorMessage(ssoError)}</span>
                                </div>
                            )}
                            <Field>
                                <FieldLabel htmlFor="email">Username</FieldLabel>
                                <Input
                                    id="email"
                                    value={email.value}
                                    onChange={email.onChange}
                                    placeholder="username or email"
                                    aria-invalid={isEmailInvalid}
                                    required
                                />
                            </Field>
                            {isEmailInvalid && <FieldError errors={[errors.email]} />}
                            <Field>
                                <div className="flex items-center justify-between">
                                    <FieldLabel htmlFor="password">Password</FieldLabel>
                                    <AppLink.Basic
                                        className="text-sm text-primary hover:text-gray-700"
                                        to={ROUTE.auth.forgotPassword.$route}
                                    >
                                        Forgot Password?
                                    </AppLink.Basic>
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
                                    Sign in
                                </Button>

                                {loginOptions.map(option => (
                                    <Button
                                        key={option.type}
                                        variant="outline"
                                        type="button"
                                        onClick={() => {
                                            void handleLoginWithProvider(option);
                                        }}
                                    >
                                        {option.icon === "github" && <GithubIcon />}
                                        {option.icon === "gitlab" && <GitlabIcon />}
                                        {option.icon === "gitea" && <GiteaIcon />}
                                        {option.icon === "google" && <GoogleIcon />}
                                        Continue with {option.name}
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
