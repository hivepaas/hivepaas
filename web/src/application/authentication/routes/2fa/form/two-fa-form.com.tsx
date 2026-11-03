import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { z } from "zod";

import { BackToSignIn } from "@application/authentication/components";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from "@/components/ui/input-otp";

const CODE_LENGTH = 6;

const Schema = z.object({
    code: z
        .string()
        .trim()
        .length(CODE_LENGTH, { message: `Code must be ${CODE_LENGTH} digits` })
        .regex(/^\d+$/, { message: "Code must contain only numbers" }),
});

type SchemaInput = z.input<typeof Schema>;
type SchemaOutput = z.output<typeof Schema>;

export function TwoFaForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            code: "",
        },
        resolver: zodResolver(Schema),
        mode: "onSubmit",
    });

    const {
        field: code,
        fieldState: { invalid: isCodeInvalid },
    } = useController({
        name: "code",
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
                    <CardTitle>Two-factor authentication</CardTitle>
                    <CardDescription>Enter the 6-digit code from your authenticator app</CardDescription>
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
                                <FieldLabel htmlFor="otp">Authentication code</FieldLabel>
                                <div className="flex justify-center">
                                    <InputOTP
                                        id="otp"
                                        value={code.value}
                                        onChange={code.onChange}
                                        maxLength={CODE_LENGTH}
                                        aria-invalid={isCodeInvalid}
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
                                {isCodeInvalid && <FieldError errors={[errors.code]} />}
                            </Field>
                            <Button
                                type="submit"
                                isLoading={isPending}
                            >
                                Verify Code
                            </Button>
                            <BackToSignIn />
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
