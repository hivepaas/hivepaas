import { useState } from "react";

import {
    Avatar,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Checkbox,
    Field,
    FieldError,
    FieldGroup,
    FieldLabel,
    Input,
} from "@components/ui";
import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from "@components/ui/input-otp";
import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { Pencil } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { PasswordStrengthMeter } from "@application/shared/components";
import { ESecuritySettings } from "@application/shared/enums";

import { BackToSignIn } from "@application/authentication/components";
import { PhotoUploadDialog } from "@application/authentication/dialogs";
import type { Candidate } from "@application/authentication/domain/user";
import {
    type SighUpFormSchemaInput,
    type SighUpFormSchemaOutput,
    SignUpFormSchema,
} from "@application/authentication/routes/sign-up/schemas";

const CODE_LENGTH = 6;

export function SignUpForm({ method, isPending, onSubmit }: Props) {
    const [openCandidatePhoto, setOpenCandidatePhoto] = useState(false);

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SighUpFormSchemaInput, unknown, SighUpFormSchemaOutput>({
        defaultValues: {
            username: "",
            fullName: "",
            email: method.candidate.email,
            password: "",
            photo: null,
            agreeTermsAndConditions: false,
            securityOption: method.candidate.securityOption,
            mfaTotpSecret:
                method.candidate.securityOption === ESecuritySettings.Password2FA ? method.candidate.mfaTotpSecret : "",
            passcode: "",
        },
        resolver: zodResolver(SignUpFormSchema),
        mode: "onSubmit",
    });

    const {
        field: username,
        fieldState: { invalid: isUsernameInvalid },
    } = useController({
        name: "username",
        control,
    });

    const {
        field: fullName,
        fieldState: { invalid: isFullNameInvalid },
    } = useController({
        name: "fullName",
        control,
    });

    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({
        name: "email",
        control,
        defaultValue: "",
    });

    const {
        field: password,
        fieldState: { invalid: isPasswordInvalid },
    } = useController({
        name: "password",
        control,
    });

    const {
        field: passcode,
        fieldState: { invalid: isPasscodeInvalid },
    } = useController({
        name: "passcode",
        control,
    });

    const { field: photo } = useController({
        name: "photo",
        control,
    });

    const { field: agreeTermsAndConditions } = useController({
        name: "agreeTermsAndConditions",
        control,
    });

    const {
        field: isStrongPassword,
        fieldState: { invalid: isNotStrongEnough },
    } = useController({
        name: "isStrongPassword",
        control,
    });

    function onValid(values: SighUpFormSchemaOutput) {
        const { photo: p } = values;

        if (p === null) {
            onSubmit(values);

            return;
        }

        const dataBase64 = p.dataBase64.split(",")[1];

        if (!dataBase64) {
            console.error("Invalid base64 data");

            return;
        }

        onSubmit({
            ...values,
            photo: { ...p, dataBase64 },
        });
    }

    function onInvalid(fieldErrors: FieldErrors<SighUpFormSchemaOutput>) {
        console.error(fieldErrors);
    }

    return (
        <>
            <div className="flex flex-col gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle>Sign up</CardTitle>
                        <CardDescription>Enter your details below to sign up</CardDescription>
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
                                    <div className="flex items-center gap-3 justify-center">
                                        <div className="relative size-24 rounded-full border">
                                            <Avatar
                                                name={fullName.value}
                                                className="size-full"
                                                src={photo.value?.dataBase64 ?? undefined}
                                            />
                                            <Button
                                                type="button"
                                                size="icon-sm"
                                                className="absolute -bottom-1 -right-1 rounded-full"
                                                onClick={() => {
                                                    setOpenCandidatePhoto(true);
                                                }}
                                                aria-label="Edit photo"
                                                title="Edit photo"
                                            >
                                                <Pencil />
                                            </Button>
                                        </div>
                                    </div>
                                    <FieldError errors={[errors.photo]} />
                                </Field>
                                <Field>
                                    <FieldLabel htmlFor="email">Email</FieldLabel>
                                    <Input
                                        id="email"
                                        value={email.value}
                                        onChange={email.onChange}
                                        type="email"
                                        placeholder="m@example.com"
                                        aria-invalid={isEmailInvalid}
                                        disabled
                                    />
                                </Field>
                                <Field>
                                    <FieldLabel
                                        htmlFor="username"
                                        isRequired
                                    >
                                        Username
                                    </FieldLabel>
                                    <Input
                                        id="username"
                                        value={username.value}
                                        onChange={username.onChange}
                                        type="text"
                                        placeholder="Enter your username"
                                        aria-invalid={isUsernameInvalid}
                                        required
                                    />
                                    <FieldError errors={[errors.username]} />
                                </Field>
                                <Field>
                                    <FieldLabel
                                        htmlFor="fullName"
                                        isRequired
                                    >
                                        Full Name
                                    </FieldLabel>
                                    <Input
                                        id="fullName"
                                        value={fullName.value}
                                        onChange={fullName.onChange}
                                        type="text"
                                        placeholder="Enter your full name"
                                        aria-invalid={isFullNameInvalid}
                                        required
                                    />
                                    <FieldError errors={[errors.fullName]} />
                                </Field>
                                <Field>
                                    <FieldLabel
                                        htmlFor="password"
                                        isRequired
                                    >
                                        Password
                                    </FieldLabel>
                                    <PasswordInput
                                        id="password"
                                        value={password.value}
                                        onChange={password.onChange}
                                        aria-invalid={isPasswordInvalid || isNotStrongEnough}
                                        required
                                    />
                                    <FieldError errors={[errors.password, errors.isStrongPassword]} />
                                </Field>

                                <Field>
                                    <PasswordStrengthMeter
                                        password={password.value}
                                        onStrengthChange={strength => {
                                            isStrongPassword.onChange(strength === "max");
                                        }}
                                    />
                                </Field>
                                {method.candidate.securityOption === ESecuritySettings.Password2FA && (
                                    <Field>
                                        <FieldLabel htmlFor="qrCode">QR Code</FieldLabel>
                                        <img
                                            src={`data:image/png;base64,${method.candidate.qrCode}`}
                                            alt="QR Code"
                                        />
                                    </Field>
                                )}

                                {method.candidate.securityOption === ESecuritySettings.Password2FA && (
                                    <Field>
                                        <FieldLabel htmlFor="passcode">Passcode</FieldLabel>
                                        <div className="flex justify-center">
                                            <InputOTP
                                                id="passcode"
                                                value={passcode.value}
                                                onChange={passcode.onChange}
                                                maxLength={CODE_LENGTH}
                                                aria-invalid={isPasscodeInvalid}
                                            >
                                                <InputOTPGroup>
                                                    <InputOTPSlot index={0} />
                                                    <InputOTPSlot index={1} />
                                                    <InputOTPSlot index={2} />
                                                </InputOTPGroup>
                                                <InputOTPSeparator />
                                                <InputOTPGroup>
                                                    <InputOTPSlot index={3} />
                                                    <InputOTPSlot index={4} />
                                                    <InputOTPSlot index={5} />
                                                </InputOTPGroup>
                                            </InputOTP>
                                        </div>
                                        <FieldError errors={[errors.passcode]} />
                                    </Field>
                                )}

                                <Field>
                                    <div className="flex items-center gap-3">
                                        <Checkbox
                                            id="agreeTermsAndConditions"
                                            checked={agreeTermsAndConditions.value}
                                            onCheckedChange={agreeTermsAndConditions.onChange}
                                        />
                                        <FieldLabel htmlFor="agreeTermsAndConditions">
                                            Agree to Terms and Conditions
                                        </FieldLabel>
                                    </div>
                                    <FieldError errors={[errors.agreeTermsAndConditions]} />
                                </Field>
                                <Field>
                                    <Button
                                        type="submit"
                                        isLoading={isPending}
                                    >
                                        Sign Up
                                    </Button>

                                    <BackToSignIn />
                                </Field>
                            </FieldGroup>
                        </form>
                    </CardContent>
                </Card>
            </div>
            <PhotoUploadDialog
                open={openCandidatePhoto}
                onOpenChange={setOpenCandidatePhoto}
                initialImage={photo.value?.dataBase64 ?? null}
                onConfirm={result => {
                    photo.onChange(result);
                }}
            />
        </>
    );
}

interface Props {
    method: { inviteToken: string; candidate: Candidate };
    isPending: boolean;
    onSubmit: (values: Omit<SighUpFormSchemaOutput, "isStrongPassword">) => void;
}
