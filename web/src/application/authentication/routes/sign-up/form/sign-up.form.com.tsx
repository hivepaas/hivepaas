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

        if (dataBase64 === undefined) {
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
                aspectRatio={1}
            />
            {/* <div className={cx("sign-up-form")}>
                <form
                    onSubmit={event => {
                        event.preventDefault();

                        void handleSubmit(onValid, onInvalid)(event);
                    }}
                >
                    <div className={cx("details")}>
                        <div className={cx("input")}>
                            <Input.Text
                                size="large"
                                label="First Name"
                                placeholder="Enter first name"
                                status={isFirstNameInvalid ? "error" : ""}
                                {...firstName}
                            />
                            <InputErrorMessage
                                name={firstName.name}
                                errors={errors}
                            />
                        </div>

                        <div className={cx("input")}>
                            <Input.Text
                                size="large"
                                label="Last Name"
                                placeholder="Enter last name"
                                status={isLastNameInvalid ? "error" : ""}
                                {...lastName}
                            />
                            <InputErrorMessage
                                name={lastName.name}
                                errors={errors}
                            />
                        </div>

                        <div className={cx("block")}>
                            <div className={cx("input")}>
                                <Input.Text
                                    size="large"
                                    label="Position"
                                    placeholder="Enter position"
                                    status={isPositionInvalid ? "error" : ""}
                                    {...position}
                                />
                                <InputErrorMessage
                                    name={position.name}
                                    errors={errors}
                                />
                            </div>

                            {method.candidate.isInternal && (
                                <div className={cx("input")}>
                                    <Select.Single
                                        ref={entity.ref}
                                        size="large"
                                        className={cx("select")}
                                        value={entity.value?.id ?? null}
                                        options={entityOptions}
                                        allowClear
                                        showSearch
                                        optionFilterProp="name"
                                        fieldNames={{
                                            value: "id",
                                            label: "name",
                                        }}
                                        onChange={(_, option) => {
                                            entity.onChange(option ?? null);
                                        }}
                                        onSearch={setEntitySearch}
                                        loading={isFetching}
                                        label="Entity"
                                        placeholder="Select entity"
                                        status={isEntityInvalid ? "error" : ""}
                                    />
                                    <InputErrorMessage
                                        name={entity.name}
                                        errors={errors}
                                    />
                                </div>
                            )}
                        </div>

                        <div className={cx("input")}>
                            <Input.Text
                                size="large"
                                className={cx("email-input")}
                                label="Email Address"
                                placeholder="Please enter your email"
                                disabled={method.type === ESignUpType.Email}
                                status={isEmailInvalid ? "error" : ""}
                                {...email}
                            />
                            <InputErrorMessage
                                name={email.name}
                                errors={errors}
                            />
                        </div>

                        <div className={cx("input")}>
                            <Select.Single
                                {...timezone}
                                className={cx("select")}
                                size="large"
                                options={timezones.map(t => {
                                    return {
                                        label: t.name,
                                        value: t.id,
                                    };
                                })}
                                allowClear
                                showSearch
                                onChange={id => {
                                    timezone.onChange(id ?? null);
                                }}
                                optionFilterProp="label"
                                label="Timezone"
                                placeholder="Select your timezone"
                                status={isTimezoneInvalid ? "error" : ""}
                            />
                            <InputErrorMessage
                                name={timezone.name}
                                errors={errors}
                            />
                        </div>

                        <div className={cx("input")}>
                            <Input.Phone
                                label="Mobile Phone"
                                size="large"
                                status={isMobilePhoneInvalid ? "error" : ""}
                                {...mobilePhone}
                            />
                            <InputErrorMessage
                                name={mobilePhone.name}
                                errors={errors}
                            />
                        </div>

                        <div className={cx("input")}>
                            <Input.Phone
                                label="Office Phone"
                                size="large"
                                status={isOfficePhoneInvalid ? "error" : ""}
                                {...officePhone}
                            />
                            <InputErrorMessage
                                name={officePhone.name}
                                errors={errors}
                            />
                        </div>
                    </div>

                    <div className={cx("input")}>
                        <Input.Password
                            size="large"
                            label="Password"
                            placeholder="Please enter your password"
                            status={isPasswordInvalid || isNotStrongEnough ? "error" : ""}
                            {...password}
                        />
                        <InputErrorMessage
                            name={isNotStrongEnough ? isStrongPassword.name : password.name}
                            errors={errors}
                        />
                    </div>

                    <div className={cx("input")}>
                        <PasswordStrengthMeter
                            password={password.value}
                            onStrengthChange={strength => {
                                isStrongPassword.onChange(strength === "max");
                            }}
                        />
                    </div>

                    <div className={cx("input")}>
                        <InputLabel>Profile Photo</InputLabel>

                        <div className={cx("profile-photo")}>
                            <Avatar.Basic
                                className={cx("avatar")}
                                borderless
                                name={`${firstName.value} ${lastName.value}`.trim() || "??"}
                                src={photo.value?.dataBase64 ?? null}
                            />

                            <Button
                                className={cx("upload-button")}
                                icon={<UploadIcon className={cx("upload-icon")} />}
                                onClick={() => {
                                    setOpenCandidatePhoto(true);
                                }}
                            >
                                {photo.value === null ? "Choose Photo" : "Change Photo"}
                            </Button>

                            <Button
                                className={cx("delete-button")}
                                icon={<TrashIcon className={cx("trash-icon")} />}
                                disabled={photo.value === null}
                                onClick={() => {
                                    photo.onChange(null);
                                }}
                            />
                        </div>

                        <InputErrorMessage
                            name={photo.name}
                            errors={errors}
                        />
                    </div>

                    <div className={cx("terms-and-conditions-input-wrapper")}>
                        <div className={cx("checkbox")}>
                            <Checkbox
                                ref={agreeTermsAndConditions.ref}
                                checked={agreeTermsAndConditions.value}
                                onChange={e => {
                                    agreeTermsAndConditions.onChange(e.target.checked);
                                }}
                            />
                            <div className={cx("title")}>
                                Agree to{" "}
                                <Button
                                    className={cx("terms-and-conditions-button")}
                                    type="link"
                                    onClick={() => {
                                        notify.warning({
                                            message: "Terms and Conditions",
                                            description: "Not implemented yet!",
                                        });
                                    }}
                                >
                                    Terms and Conditions
                                </Button>
                            </div>
                        </div>

                        <InputErrorMessage
                            name={agreeTermsAndConditions.name}
                            errors={errors}
                        />
                    </div>

                    <div className={cx("submit-button-input-wrapper")}>
                        <Button
                            size="large"
                            type="primary"
                            htmlType="submit"
                            loading={isPending}
                        >
                            Sign Up
                        </Button>
                    </div>

                    <div className={cx("to-sign-in-input-wrapper")}>
                        <span>Already have an account?</span>
                        <AppLink.Basic
                            className={cx("sign-in-link")}
                            to={ROUTE.auth.signIn.$route}
                        >
                            Sign In
                        </AppLink.Basic>
                    </div>
                </form>
            </div>

            <CandidatePhotoModal
                photo={photo.value ? photo.value.dataBase64 : null}
                open={openCandidatePhoto}
                onClose={() => {
                    setOpenCandidatePhoto(false);
                }}
                onChange={photo.onChange}
            /> */}
        </>
    );
}

interface Props {
    method: { inviteToken: string; candidate: Candidate };
    isPending: boolean;
    onSubmit: (values: Omit<SighUpFormSchemaOutput, "isStrongPassword">) => void;
}
