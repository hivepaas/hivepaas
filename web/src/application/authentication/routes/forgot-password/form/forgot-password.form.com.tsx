import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { z } from "zod";

import { BackToSignIn } from "@application/authentication/components";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

const Schema = z.object({
    email: z.string().trim().email(),
});

type SchemaInput = z.input<typeof Schema>;
type SchemaOutput = z.output<typeof Schema>;

export function ForgotPasswordForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            email: "",
        },
        resolver: zodResolver(Schema),
        mode: "onSubmit",
    });

    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({
        name: "email",
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
                    <CardTitle>Forgot password</CardTitle>
                    <CardDescription>Enter your email to receive a reset link</CardDescription>
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
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                >
                                    Send reset link
                                </Button>
                                <BackToSignIn />
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
    onSubmit: (values: SchemaOutput) => Promise<void> | void;
}
