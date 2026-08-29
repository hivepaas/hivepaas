import { useEffect } from "react";

import { DialogActionFooter, DialogBody } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useForm } from "react-hook-form";

import { Button, Field, FieldError, Input } from "@/components/ui";

import {
    type ConfirmEnvDangerActionFormInput,
    type ConfirmEnvDangerActionFormOutput,
    createConfirmEnvDangerActionFormSchema,
} from "../schemas";
import { EnvDangerAction } from "../types";

const actionCopy = {
    [EnvDangerAction.Disable]: {
        bodyAction: "disabling",
        buttonLabel: "Disable this Environment",
        buttonVariant: "destructive",
    },
    [EnvDangerAction.ReEnable]: {
        bodyAction: "re-enabling",
        buttonLabel: "Re-enable this Environment",
        buttonVariant: "default",
    },
    [EnvDangerAction.Delete]: {
        bodyAction: "deleting",
        buttonLabel: "Delete this Environment",
        buttonVariant: "destructive",
    },
} as const;

export function ConfirmEnvDangerActionForm({ action, envName, isPending = false, readOnly = false, onSubmit }: Props) {
    const {
        formState: { errors },
        handleSubmit,
        register,
        reset,
        watch,
    } = useForm<ConfirmEnvDangerActionFormInput, unknown, ConfirmEnvDangerActionFormOutput>({
        defaultValues: {
            envName: "",
        },
        resolver: zodResolver(createConfirmEnvDangerActionFormSchema(envName)),
        mode: "onSubmit",
    });

    useEffect(() => {
        reset({
            envName: "",
        });
    }, [action, envName, reset]);

    const enteredEnvName = watch("envName");
    const isConfirmed = enteredEnvName === envName;
    const copy = actionCopy[action];

    function onInvalid(_errors: FieldErrors<ConfirmEnvDangerActionFormInput>) {
        return undefined;
    }

    return (
        <form
            className="min-h-0 flex flex-1 flex-col"
            onSubmit={event => {
                event.preventDefault();
                if (readOnly) {
                    return;
                }

                void handleSubmit(onSubmit, onInvalid)(event);
            }}
        >
            <DialogBody className="flex flex-col gap-7">
                <p className="text-sm font-medium leading-6 text-foreground">
                    To confirm {copy.bodyAction} the environment, please type{" "}
                    <span className="inline-flex max-w-full items-center rounded-[4px] border border-input bg-muted/50 px-2 py-0 align-middle font-semibold text-primary break-all">
                        {envName}
                    </span>{" "}
                    into the text box below.
                </p>

                <Field>
                    <Input
                        aria-invalid={Boolean(errors.envName)}
                        disabled={readOnly || isPending}
                        {...register("envName")}
                    />
                    <FieldError errors={[errors.envName]} />
                </Field>
            </DialogBody>

            <DialogActionFooter>
                <Button
                    type="submit"
                    variant={copy.buttonVariant}
                    disabled={readOnly || isPending || !isConfirmed}
                    isLoading={isPending}
                    className="min-w-[120px]"
                >
                    {copy.buttonLabel}
                </Button>
            </DialogActionFooter>
        </form>
    );
}

interface Props {
    action: EnvDangerAction;
    envName: string;
    isPending?: boolean;
    readOnly?: boolean;
    onSubmit: (values: ConfirmEnvDangerActionFormOutput) => void;
}
