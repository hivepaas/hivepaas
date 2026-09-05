import { InputOTP, InputOTPGroup, InputOTPSlot } from "@components/ui/input-otp";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import z from "zod";

import { MfaQrCode } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";

const CODE_LENGTH = 6;

export const F2aSetupSchema = z.object({
    passcode: z.string().trim().min(1, "Passcode is required"),
});

export type F2aSetupSchemaInput = z.input<typeof F2aSetupSchema>;
export type F2aSetupSchemaOutput = z.output<typeof F2aSetupSchema>;

export function F2aSetupForm({ isPending, onSubmit, qrCode, totpToken, secretKey }: Props) {
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
        <form
            onSubmit={event => {
                event.preventDefault();

                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-6">
                <FieldGroup>
                    <Field>
                        <FieldLabel>Scan this QR code with your authenticator app</FieldLabel>
                        <MfaQrCode
                            qrCode={qrCode}
                            secretKey={secretKey}
                        />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="passcode">Then, enter the passcode here</FieldLabel>
                        <div className="flex justify-center">
                            <InputOTP
                                id="passcode"
                                value={passcode.value}
                                onChange={passcode.onChange}
                                maxLength={CODE_LENGTH}
                                aria-invalid={isPasscodeInvalid}
                            >
                                <InputOTPGroup className="gap-2 sm:gap-2.5">
                                    <InputOTPSlot
                                        index={0}
                                        className="size-10 sm:size-11 rounded-md border border-input dark:border-white/25 dark:bg-white/5 text-base sm:text-lg font-semibold shadow-xs"
                                    />
                                    <InputOTPSlot
                                        index={1}
                                        className="size-10 sm:size-11 rounded-md border border-input dark:border-white/25 dark:bg-white/5 text-base sm:text-lg font-semibold shadow-xs"
                                    />
                                    <InputOTPSlot
                                        index={2}
                                        className="size-10 sm:size-11 rounded-md border border-input dark:border-white/25 dark:bg-white/5 text-base sm:text-lg font-semibold shadow-xs"
                                    />
                                    <InputOTPSlot
                                        index={3}
                                        className="size-10 sm:size-11 rounded-md border border-input dark:border-white/25 dark:bg-white/5 text-base sm:text-lg font-semibold shadow-xs"
                                    />
                                    <InputOTPSlot
                                        index={4}
                                        className="size-10 sm:size-11 rounded-md border border-input dark:border-white/25 dark:bg-white/5 text-base sm:text-lg font-semibold shadow-xs"
                                    />
                                    <InputOTPSlot
                                        index={5}
                                        className="size-10 sm:size-11 rounded-md border border-input dark:border-white/25 dark:bg-white/5 text-base sm:text-lg font-semibold shadow-xs"
                                    />
                                </InputOTPGroup>
                            </InputOTP>
                        </div>
                        <FieldError errors={[errors.passcode]} />
                    </Field>
                </FieldGroup>
            </DialogBody>
            <DialogActionFooter>
                <Button
                    type="submit"
                    isLoading={isPending}
                    className="min-w-[100px]"
                >
                    Activate
                </Button>
            </DialogActionFooter>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: F2aSetupSchemaOutput) => Promise<void> | void;
    qrCode: string;
    totpToken: string;
    secretKey: string;
}
