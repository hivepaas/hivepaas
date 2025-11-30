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
import { ImageService } from "@infrastructure/services";
import { Link, Pencil } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { MfaQrCode, PasswordStrengthMeter } from "@application/shared/components";
import { PhotoUploadDialog } from "@application/shared/dialogs";
import { ESecuritySettings } from "@application/shared/enums";

import { BackToSignIn } from "@application/authentication/components";
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
        setValue,
        control,
        formState: { errors },
    } = useForm<SighUpFormSchemaInput, unknown, SighUpFormSchemaOutput>({
        defaultValues: {
            username: "",
            fullName: "",
            email: method.candidate.email,
            password: "",
            confirmPassword: "",
            position: "",
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

    const {
        field: confirmPassword,
        fieldState: { invalid: isConfirmPasswordInvalid },
    } = useController({
        name: "confirmPassword",
        control,
    });

    const {
        field: position,
        fieldState: { invalid: isPositionInvalid },
    } = useController({
        name: "position",
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

        onSubmit({
            ...values,
            photo: p,
        });
    }

    function onInvalid(fieldErrors: FieldErrors<SighUpFormSchemaOutput>) {
        console.error(fieldErrors);
    }

    async function handlePhotoUpload(result: File | null) {
        if (!result) {
            photo.onChange(null);
            return;
        }

        const base64String = await ImageService.convertFileToBase64(result);

        setValue("photo", {
            fileName: result.name,
            dataBase64: base64String,
        });
    }
    return (
        <>
            <div className="flex flex-col gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle className="text-[30px]">Sign up</CardTitle>
                        <CardDescription>Please fill your details to complete registration process</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form
                            onSubmit={event => {
                                event.preventDefault();

                                void handleSubmit(onValid, onInvalid)(event);
                            }}
                        >
                            <FieldGroup className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <Field className="col-span-full md:col-span-1">
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
                                <Field className="col-span-full md:col-span-1">
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
                                <Field className="col-span-full md:col-span-1">
                                    <FieldLabel htmlFor="position">Position</FieldLabel>
                                    <Input
                                        id="position"
                                        value={position.value}
                                        onChange={position.onChange}
                                        type="text"
                                        placeholder="Enter your position"
                                        aria-invalid={isPositionInvalid}
                                    />
                                    <FieldError errors={[errors.position]} />
                                </Field>
                                <Field className="col-span-full md:col-span-1">
                                    <FieldLabel htmlFor="email">Email address</FieldLabel>
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

                                <Field className="col-span-full md:col-span-1">
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
                                <Field className="col-span-full md:col-span-1">
                                    <FieldLabel
                                        htmlFor="confirmPassword"
                                        isRequired
                                    >
                                        Retype Password
                                    </FieldLabel>
                                    <PasswordInput
                                        id="confirmPassword"
                                        value={confirmPassword.value}
                                        onChange={confirmPassword.onChange}
                                        aria-invalid={isConfirmPasswordInvalid}
                                        required
                                    />
                                    <FieldError errors={[errors.confirmPassword]} />
                                </Field>

                                <Field className="col-span-full">
                                    <PasswordStrengthMeter
                                        password={password.value}
                                        onStrengthChange={strength => {
                                            isStrongPassword.onChange(strength === "max");
                                        }}
                                    />
                                </Field>

                                <Field className="col-span-full md:col-span-1">
                                    <FieldLabel htmlFor="photo">Profile Photo</FieldLabel>
                                    <div className="flex items-center gap-3">
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

                                {method.candidate.securityOption === ESecuritySettings.Password2FA && (
                                    <Field className="col-span-full md:col-span-1">
                                        <FieldLabel>Setup 2-Factor Authentication</FieldLabel>
                                        <MfaQrCode
                                            qrCode={method.candidate.qrCode}
                                            secretKey={method.candidate.mfaTotpSecret}
                                        />
                                        <FieldLabel htmlFor="passcode">Enter the generated passcode</FieldLabel>

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
                                                </InputOTPGroup>
                                                <InputOTPSeparator />
                                                <InputOTPGroup>
                                                    <InputOTPSlot index={2} />
                                                    <InputOTPSlot index={3} />
                                                </InputOTPGroup>
                                                <InputOTPSeparator />
                                                <InputOTPGroup>
                                                    <InputOTPSlot index={4} />
                                                    <InputOTPSlot index={5} />
                                                </InputOTPGroup>
                                            </InputOTP>
                                        </div>
                                        <FieldError errors={[errors.passcode]} />
                                    </Field>
                                )}
                                <Field className="col-span-full">
                                    <div className="flex items-center gap-3">
                                        <Checkbox
                                            id="agreeTermsAndConditions"
                                            checked={agreeTermsAndConditions.value}
                                            onCheckedChange={agreeTermsAndConditions.onChange}
                                        />
                                        <FieldLabel
                                            htmlFor="agreeTermsAndConditions"
                                            className="font-normal"
                                        >
                                            Agree to{" "}
                                            <a
                                                href="/terms-and-conditions"
                                                className="text-primary underline"
                                            >
                                                Terms and Conditions
                                            </a>
                                        </FieldLabel>
                                    </div>
                                    <FieldError errors={[errors.agreeTermsAndConditions]} />
                                </Field>
                                <Field className="col-span-full">
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
                onSubmit={result => {
                    void handlePhotoUpload(result);
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
