import React, { useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

import type {
    CreateOrEditProjectSecretFormInput,
    CreateOrEditProjectSecretFormOutput,
} from "../schemas";
import { CreateOrEditProjectSecretFormSchema } from "../schemas";

export function CreateOrEditProjectSecretForm({ isPending, onSubmit, onHasChanges, isEditMode, initialValues }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditProjectSecretFormInput, unknown, CreateOrEditProjectSecretFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            value: "",
        },
        resolver: zodResolver(CreateOrEditProjectSecretFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        name: "name",
        control,
    });

    const {
        field: value,
        fieldState: { invalid: isValueInvalid },
    } = useController({
        name: "value",
        control,
    });

    function onValid(values: CreateOrEditProjectSecretFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditProjectSecretFormOutput>) {
        console.error(_errors);
        // Optional: log errors or show notification
    }

    return (
        <div className="flex flex-col gap-6">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="flex flex-col gap-6"
            >
                <InfoBlock
                    titleWidth={150}
                    title={
                        <LabelWithInfo
                            label="Name"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <Input
                                id="name"
                                {...name}
                                placeholder="Enter secret name"
                                aria-invalid={isNameInvalid}
                                disabled={isEditMode}
                            />
                            <FieldError errors={[errors.name]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={150}
                    title={
                        <LabelWithInfo
                            label="Value"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <Textarea
                                id="value"
                                {...value}
                                placeholder="Enter secret value"
                                rows={6}
                                aria-invalid={isValueInvalid}
                            />
                            <FieldError errors={[errors.value]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <Field>
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            {isEditMode ? "Update Secret" : "Create Secret"}
                        </Button>
                    </div>
                </Field>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditProjectSecretFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    isEditMode: boolean;
    initialValues?: CreateOrEditProjectSecretFormInput;
}
