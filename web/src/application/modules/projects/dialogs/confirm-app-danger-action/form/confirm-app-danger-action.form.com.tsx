import { useEffect } from "react";

import { DialogFooter } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useForm } from "react-hook-form";

import { Button, Field, FieldError, Input } from "@/components/ui";

import {
    type ConfirmAppDangerActionFormInput,
    type ConfirmAppDangerActionFormOutput,
    createConfirmAppDangerActionFormSchema,
} from "../schemas";
import { AppDangerAction } from "../types";

const actionCopy = {
    [AppDangerAction.Disable]: {
        bodyAction: "disabling",
        buttonLabel: "Disable this App",
        buttonVariant: "destructive",
    },
    [AppDangerAction.ReEnable]: {
        bodyAction: "re-enabling",
        buttonLabel: "Re-enable this App",
        buttonVariant: "default",
    },
    [AppDangerAction.Delete]: {
        bodyAction: "deleting",
        buttonLabel: "Delete this App",
        buttonVariant: "destructive",
    },
} as const;

export function ConfirmAppDangerActionForm({ action, appName, isPending = false, readOnly = false, onSubmit }: Props) {
    const {
        formState: { errors },
        handleSubmit,
        register,
        reset,
        watch,
    } = useForm<ConfirmAppDangerActionFormInput, unknown, ConfirmAppDangerActionFormOutput>({
        defaultValues: {
            appName: "",
        },
        resolver: zodResolver(createConfirmAppDangerActionFormSchema(appName)),
        mode: "onSubmit",
    });

    useEffect(() => {
        reset({
            appName: "",
        });
    }, [action, appName, reset]);

    const enteredAppName = watch("appName");
    const isConfirmed = enteredAppName === appName;
    const copy = actionCopy[action];

    function onInvalid(_errors: FieldErrors<ConfirmAppDangerActionFormInput>) {
        return undefined;
    }

    return (
        <form
            className="flex flex-col gap-7"
            onSubmit={event => {
                event.preventDefault();
                if (readOnly) {
                    return;
                }

                void handleSubmit(onSubmit, onInvalid)(event);
            }}
        >
            <p className="text-sm font-medium leading-6 text-foreground">
                To confirm {copy.bodyAction} the application, please type{" "}
                <span className="text-primary">&quot;{appName}&quot;</span> into the text box below.
            </p>

            <Field>
                <Input
                    aria-invalid={Boolean(errors.appName)}
                    disabled={readOnly || isPending}
                    {...register("appName")}
                />
                <FieldError errors={[errors.appName]} />
            </Field>

            <DialogFooter>
                <Button
                    type="submit"
                    variant={copy.buttonVariant}
                    disabled={readOnly || isPending || !isConfirmed}
                    isLoading={isPending}
                    className="min-w-[160px]"
                >
                    {copy.buttonLabel}
                </Button>
            </DialogFooter>
        </form>
    );
}

interface Props {
    action: AppDangerAction;
    appName: string;
    isPending?: boolean;
    readOnly?: boolean;
    onSubmit: (values: ConfirmAppDangerActionFormOutput) => void;
}
