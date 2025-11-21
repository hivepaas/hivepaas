import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from "@components/ui/input-otp";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import z from "zod";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";

const CODE_LENGTH = 6;

export const F2aSetupSchema = z.object({
    passcode: z.string().trim().min(1, "Passcode is required"),
});

export type F2aSetupSchemaInput = z.input<typeof F2aSetupSchema>;
export type F2aSetupSchemaOutput = z.output<typeof F2aSetupSchema>;

export function F2aSetupForm({ isPending, onSubmit, qrCode, totpToken }: Props) {
    void totpToken;
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<F2aSetupSchemaInput, unknown, F2aSetupSchemaOutput>({
        defaultValues: {
            passcode: "",
        },
        resolver: zodResolver(F2aSetupSchema),
        mode: "onSubmit",
    });

    const {
        field: passcode,
        fieldState: { invalid: isPasscodeInvalid },
    } = useController({
        name: "passcode",
        control,
    });

    function onValid(values: F2aSetupSchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<F2aSetupSchemaOutput>) {
        console.log(_errors);
    }

    return (
        <div className="flex flex-col gap-6">
            <form
                onSubmit={event => {
                    event.preventDefault();

                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="qrCode">QR Code</FieldLabel>
                        <img
                            src={`data:image/png;base64,${qrCode}`}
                            alt="QR Code"
                            className="w-[220px] h-[220px] object-contain"
                        />
                    </Field>

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

                    <Field>
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Verify
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: F2aSetupSchemaOutput) => Promise<void> | void;
    qrCode: string;
    totpToken: string;
}
