import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from "@components/ui/input-otp";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import z from "zod";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";

const CODE_LENGTH = 6;

export const CurrentPasscodeSchema = z.object({
    currentPasscode: z.string().trim().min(1, "Passcode is required"),
});

export type CurrentPasscodeSchemaInput = z.input<typeof CurrentPasscodeSchema>;
export type CurrentPasscodeSchemaOutput = z.output<typeof CurrentPasscodeSchema>;

export function CurrentPasscodeForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<CurrentPasscodeSchemaInput, unknown, CurrentPasscodeSchemaOutput>({
        defaultValues: {
            currentPasscode: "",
        },
        resolver: zodResolver(CurrentPasscodeSchema),
        mode: "onSubmit",
    });

    const {
        field: currentPasscode,
        fieldState: { invalid: isCurrentPasscodeInvalid },
    } = useController({
        name: "currentPasscode",
        control,
    });

    function onValid(values: CurrentPasscodeSchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CurrentPasscodeSchemaOutput>) {
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
                        <FieldLabel htmlFor="currentPasscode">Current Passcode</FieldLabel>
                        <div className="flex justify-center">
                            <InputOTP
                                id="currentPasscode"
                                value={currentPasscode.value}
                                onChange={currentPasscode.onChange}
                                maxLength={CODE_LENGTH}
                                aria-invalid={isCurrentPasscodeInvalid}
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
                        <FieldError errors={[errors.currentPasscode]} />
                    </Field>

                    <Field>
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Continue
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CurrentPasscodeSchemaOutput) => Promise<void> | void;
}

