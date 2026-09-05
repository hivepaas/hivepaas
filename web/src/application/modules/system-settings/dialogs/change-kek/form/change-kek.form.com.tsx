import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";

import { PasswordStrengthMeter } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";

import { ChangeKekFormSchema, type ChangeKekFormSchemaInput, type ChangeKekFormSchemaOutput } from "../schemas";

export function ChangeKekForm({ isPending, onSubmit, onHasChanges }: Props) {
    const {
        handleSubmit,
        control,
        watch,
        formState: { errors },
    } = useForm<ChangeKekFormSchemaInput, unknown, ChangeKekFormSchemaOutput>({
        defaultValues: {
            currentSecret: "",
            newSecret: "",
            confirmNewSecret: "",
            isSaved: false,
            isStrongSecret: false,
        },
        resolver: zodResolver(ChangeKekFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty]);

    const {
        field: currentSecret,
        fieldState: { invalid: isCurrentSecretInvalid },
    } = useController({
        name: "currentSecret",
        control,
    });

    const {
        field: newSecret,
        fieldState: { invalid: isNewSecretInvalid },
    } = useController({
        name: "newSecret",
        control,
    });

    const {
        field: confirmNewSecret,
        fieldState: { invalid: isConfirmNewSecretInvalid },
    } = useController({
        name: "confirmNewSecret",
        control,
    });

    const {
        field: isStrongSecret,
        fieldState: { invalid: isNotStrongEnough },
    } = useController({
        name: "isStrongSecret",
        control,
    });

    const { field: isSaved } = useController({
        name: "isSaved",
        control,
    });

    const isSavedValue = watch("isSaved");

    function onValid(values: ChangeKekFormSchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(fieldErrors: FieldErrors<ChangeKekFormSchemaOutput>) {
        console.error(fieldErrors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();

                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-4">
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="currentSecret">Current Secret</FieldLabel>
                        <PasswordInput
                            id="currentSecret"
                            placeholder="Current secret"
                            value={currentSecret.value}
                            onChange={currentSecret.onChange}
                            aria-invalid={isCurrentSecretInvalid}
                        />
                        <FieldError errors={[errors.currentSecret]} />
                    </Field>
                    <Field>
                        <FieldLabel htmlFor="newSecret">New Secret</FieldLabel>
                        <PasswordInput
                            id="newSecret"
                            placeholder="New secret"
                            value={newSecret.value}
                            onChange={newSecret.onChange}
                            aria-invalid={isNewSecretInvalid || isNotStrongEnough}
                        />
                        <FieldError errors={[errors.newSecret, errors.isStrongSecret]} />
                    </Field>
                    <Field>
                        <FieldLabel htmlFor="confirmNewSecret">Confirm New Secret</FieldLabel>
                        <PasswordInput
                            id="confirmNewSecret"
                            placeholder="Confirm new secret"
                            value={confirmNewSecret.value}
                            onChange={confirmNewSecret.onChange}
                            aria-invalid={isConfirmNewSecretInvalid || isNotStrongEnough}
                        />
                        <FieldError errors={[errors.confirmNewSecret, errors.isStrongSecret]} />
                    </Field>
                    <Field>
                        <PasswordStrengthMeter
                            password={newSecret.value}
                            minLength={12}
                            onStrengthChange={strength => {
                                isStrongSecret.onChange(strength === "max");
                            }}
                        />
                    </Field>
                    <div className="flex items-center gap-2 pt-2">
                        <Checkbox
                            id="isSavedSecret"
                            checked={isSaved.value}
                            onCheckedChange={checked => {
                                isSaved.onChange(checked === true);
                            }}
                        />
                        <label
                            htmlFor="isSavedSecret"
                            className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer select-none text-destructive"
                        >
                            I confirm I have saved the secret
                        </label>
                    </div>
                </FieldGroup>
            </DialogBody>
            <DialogActionFooter>
                <Button
                    type="submit"
                    isLoading={isPending}
                    disabled={!isSavedValue || isPending}
                    className="min-w-[100px]"
                >
                    Change
                </Button>
            </DialogActionFooter>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: ChangeKekFormSchemaOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}
