import { useEffect } from "react";

import { DialogFooter } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useForm } from "react-hook-form";

import { Button, Field, FieldError, Input } from "@/components/ui";

import {
    type ConfirmProjectDangerActionFormInput,
    type ConfirmProjectDangerActionFormOutput,
    createConfirmProjectDangerActionFormSchema,
} from "../schemas";
import { ProjectDangerAction } from "../types";

const actionCopy = {
    [ProjectDangerAction.Disable]: {
        bodyAction: "disabling",
        buttonLabel: "Disable this Project",
        buttonVariant: "destructive",
    },
    [ProjectDangerAction.ReEnable]: {
        bodyAction: "re-enabling",
        buttonLabel: "Re-enable this Project",
        buttonVariant: "default",
    },
    [ProjectDangerAction.Delete]: {
        bodyAction: "deleting",
        buttonLabel: "Delete this Project",
        buttonVariant: "destructive",
    },
} as const;

export function ConfirmProjectDangerActionForm({
    action,
    projectName,
    isPending = false,
    readOnly = false,
    onSubmit,
}: Props) {
    const {
        formState: { errors },
        handleSubmit,
        register,
        reset,
        watch,
    } = useForm<ConfirmProjectDangerActionFormInput, unknown, ConfirmProjectDangerActionFormOutput>({
        defaultValues: {
            projectName: "",
        },
        resolver: zodResolver(createConfirmProjectDangerActionFormSchema(projectName)),
        mode: "onSubmit",
    });

    useEffect(() => {
        reset({
            projectName: "",
        });
    }, [action, projectName, reset]);

    const enteredProjectName = watch("projectName");
    const isConfirmed = enteredProjectName === projectName;
    const copy = actionCopy[action];

    function onInvalid(_errors: FieldErrors<ConfirmProjectDangerActionFormInput>) {
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
                To confirm {copy.bodyAction} the project, please type{" "}
                <span className="text-primary">&quot;{projectName}&quot;</span> into the text box below.
            </p>

            <Field>
                <Input
                    aria-invalid={Boolean(errors.projectName)}
                    disabled={readOnly || isPending}
                    {...register("projectName")}
                />
                <FieldError errors={[errors.projectName]} />
            </Field>

            <DialogFooter>
                <Button
                    type="submit"
                    variant={copy.buttonVariant}
                    disabled={readOnly || isPending || !isConfirmed}
                    isLoading={isPending}
                    className="min-w-[180px]"
                >
                    {copy.buttonLabel}
                </Button>
            </DialogFooter>
        </form>
    );
}

interface Props {
    action: ProjectDangerAction;
    projectName: string;
    isPending?: boolean;
    readOnly?: boolean;
    onSubmit: (values: ConfirmProjectDangerActionFormOutput) => void;
}
