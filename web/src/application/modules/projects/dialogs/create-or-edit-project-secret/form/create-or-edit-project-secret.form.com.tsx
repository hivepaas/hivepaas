import React, { useEffect, useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { UploadIcon } from "lucide-react";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditProjectSecretFormInput, CreateOrEditProjectSecretFormOutput } from "../schemas";
import { CreateOrEditProjectSecretFormSchema } from "../schemas";

export function CreateOrEditProjectSecretForm({ isPending, onSubmit, onHasChanges, isEditMode, initialValues }: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditProjectSecretFormInput, unknown, CreateOrEditProjectSecretFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            valueType: initialValues?.valueType ?? "text",
            isEditMode,
            textValue: "",
            binaryFile: null,
        },
        resolver: zodResolver(CreateOrEditProjectSecretFormSchema),
        mode: "onSubmit",
    });

    const valueType = useWatch({ control, name: "valueType" });
    const selectedFile = useWatch({ control, name: "binaryFile" });

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
        field: textValue,
        fieldState: { invalid: isTextValueInvalid },
    } = useController({
        name: "textValue",
        control,
    });

    const { field: valueTypeField } = useController({
        name: "valueType",
        control,
    });

    const { field: binaryFileField } = useController({
        name: "binaryFile",
        control,
    });

    function onValid(values: CreateOrEditProjectSecretFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditProjectSecretFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="flex flex-col gap-6"
        >
            <InfoBlock
                titleWidth={220}
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
                            id="project-secret-name"
                            {...name}
                            placeholder="SECRET_NAME"
                            aria-invalid={isNameInvalid}
                            disabled={isEditMode}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Value Type"
                        isRequired
                    />
                }
            >
                <Tabs
                    value={valueType}
                    onValueChange={nextValue => {
                        valueTypeField.onChange(nextValue);
                    }}
                >
                    <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                        <TabsTrigger value="text">Text</TabsTrigger>
                        <TabsTrigger value="binary">Binary</TabsTrigger>
                    </TabsList>
                </Tabs>
            </InfoBlock>

            {valueType === "text" ? (
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Value"
                            isRequired={!isEditMode}
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <Textarea
                                id="project-secret-text-value"
                                {...textValue}
                                placeholder={isEditMode ? "Leave empty to keep current value" : "Enter secret value"}
                                rows={8}
                                aria-invalid={isTextValueInvalid}
                            />
                            <p className="text-sm text-muted-foreground">Max size: 500kb</p>
                            <FieldError errors={[errors.textValue]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>
            ) : (
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Value"
                            isRequired={!isEditMode}
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <div className="flex items-center gap-3">
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        fileInputRef.current?.click();
                                    }}
                                >
                                    <UploadIcon className="size-4" />
                                    Choose File
                                </Button>
                                <span className="truncate text-sm text-muted-foreground">
                                    {selectedFile?.name ?? (isEditMode ? "Leave empty to keep current value" : "")}
                                </span>
                            </div>
                            <Input
                                id="project-secret-binary-value"
                                ref={fileInputRef}
                                type="file"
                                className="hidden"
                                onChange={event => {
                                    binaryFileField.onChange(event.target.files?.[0] ?? null);
                                }}
                            />
                            <p className="text-sm text-muted-foreground">Max size: 500kb</p>
                            <FieldError errors={[errors.binaryFile]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>
            )}

            <Field>
                <div className="flex justify-end">
                    <Button
                        type="submit"
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </Field>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditProjectSecretFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    isEditMode: boolean;
    initialValues?: Partial<CreateOrEditProjectSecretFormInput>;
}
