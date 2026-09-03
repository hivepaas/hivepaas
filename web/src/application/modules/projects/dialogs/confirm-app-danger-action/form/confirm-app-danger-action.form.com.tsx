import { useEffect } from "react";

import { cn } from "@/lib/utils";
import { DialogActionFooter, DialogBody } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { AlertTriangle } from "lucide-react";
import { type FieldErrors, useForm } from "react-hook-form";

import { Button, Field, FieldError, Input } from "@/components/ui";

import {
    type ConfirmAppDangerActionFormInput,
    type ConfirmAppDangerActionFormOutput,
    createConfirmAppDangerActionFormSchema,
} from "../schemas";
import { AppDangerAction } from "../types";

const actionCopy: Record<
    AppDangerAction,
    {
        bodyAction: string;
        buttonLabel: string;
        buttonVariant: "default" | "destructive";
        warning?: string;
    }
> = {
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
        warning: "This action is permanent and cannot be undone.",
    },
};

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
                    To confirm {copy.bodyAction} the application, please type{" "}
                    <span className="inline-flex max-w-full items-center rounded-[4px] border border-input bg-muted/50 px-2 py-0 align-middle font-semibold text-primary break-all">
                        {appName}
                    </span>{" "}
                    into the text box below.
                </p>

                {copy.warning && (
                    <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                        <AlertTriangle className="size-4 shrink-0 text-destructive" />
                        <span>Warning: {copy.warning}</span>
                    </div>
                )}

                <Field>
                    <Input
                        aria-invalid={Boolean(errors.appName)}
                        disabled={readOnly || isPending}
                        {...register("appName")}
                    />
                    <FieldError errors={[errors.appName]} />
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
    action: AppDangerAction;
    appName: string;
    isPending?: boolean;
    readOnly?: boolean;
    onSubmit: (values: ConfirmAppDangerActionFormOutput) => void;
}
