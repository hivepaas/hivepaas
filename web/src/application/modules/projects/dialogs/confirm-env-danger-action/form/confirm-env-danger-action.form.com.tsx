import { useEffect } from "react";

import { cn } from "@/lib/utils";
import { DialogActionFooter, DialogBody } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { AlertTriangle } from "lucide-react";
import { type FieldErrors, useForm } from "react-hook-form";
import { ConfirmDangerTargetBadge } from "~/projects/module-shared/components";

import { Button, Checkbox, Field, FieldError, Input } from "@/components/ui";

import {
    type ConfirmEnvDangerActionFormInput,
    type ConfirmEnvDangerActionFormOutput,
    createConfirmEnvDangerActionFormSchema,
} from "../schemas";
import { EnvDangerAction } from "../types";

const actionCopy: Record<
    EnvDangerAction,
    {
        bodyAction: string;
        buttonLabel: string;
        buttonVariant: "default" | "destructive";
        warning?: string;
    }
> = {
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
        warning: "This action is permanent and cannot be undone.",
    },
};

export function ConfirmEnvDangerActionForm({ action, envName, isPending = false, readOnly = false, onSubmit }: Props) {
    const {
        formState: { errors },
        handleSubmit,
        register,
        reset,
        setValue,
        watch,
    } = useForm<ConfirmEnvDangerActionFormInput, unknown, ConfirmEnvDangerActionFormOutput>({
        defaultValues: {
            envName: "",
            removeStorage: false,
        },
        resolver: zodResolver(createConfirmEnvDangerActionFormSchema(envName)),
        mode: "onSubmit",
    });

    useEffect(() => {
        reset({
            envName: "",
            removeStorage: false,
        });
    }, [action, envName, reset]);

    const enteredEnvName = watch("envName");
    const isConfirmed = enteredEnvName === envName;
    const removeStorage = watch("removeStorage");
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
            <DialogBody className={cn("flex flex-col", copy.warning ? "gap-4" : "gap-7")}>
                <p className="text-sm font-medium leading-6 text-foreground">
                    To confirm {copy.bodyAction} the environment, please type{" "}
                    <ConfirmDangerTargetBadge text={envName} /> into the text box below.
                </p>

                {copy.warning && (
                    <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                        <AlertTriangle className="size-4 shrink-0 text-destructive" />
                        <span>Warning: {copy.warning}</span>
                    </div>
                )}

                {action === EnvDangerAction.Delete && (
                    <label
                        className="flex items-start gap-2.5 text-sm leading-6 text-foreground"
                        htmlFor="env-remove-storage"
                    >
                        <Checkbox
                            id="env-remove-storage"
                            className="mt-1"
                            checked={removeStorage}
                            disabled={readOnly || isPending}
                            onCheckedChange={checked => {
                                setValue("removeStorage", checked === true);
                            }}
                        />
                        <span>
                            <span className="block">Also delete the stored data</span>
                            <span className="block text-muted-foreground">
                                The volumes this environment owns, and what is on them, are kept unless you tick this.
                            </span>
                        </span>
                    </label>
                )}

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
