import React, { useEffect, useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Download, UploadIcon } from "lucide-react";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditProjectSecretFormInput, CreateOrEditProjectSecretFormOutput } from "../schemas";
import { CreateOrEditProjectSecretFormSchema } from "../schemas";

export function CreateOrEditProjectSecretForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    revealedSecret,
    revealedVersion,
    isEditMode,
    initialValues,
    readOnly = false,
    stickyActions = false,
    onClose,
}: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    const {
        handleSubmit,
        control,
        getValues,
        setValue,
        reset,
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
        if (savedVersion === 0) {
            return;
        }

        reset(getValues());
        onHasChanges?.(false);
    }, [getValues, onHasChanges, reset, savedVersion]);

    useEffect(() => {
        if (revealedSecret !== undefined && revealedSecret !== null) {
            setValue("textValue", revealedSecret, { shouldDirty: false });
            if (initialValues?.valueType === "binary" && revealedSecret) {
                try {
                    const binaryStr = window.atob(revealedSecret);
                    const bytes = new Uint8Array(binaryStr.length);
                    for (let i = 0; i < binaryStr.length; i++) {
                        bytes[i] = binaryStr.charCodeAt(i);
                    }
                    const fileName = initialValues.name ? `${initialValues.name}.bin` : "secret.bin";
                    const file = new File([bytes], fileName);
                    setValue("binaryFile", file, { shouldDirty: false });
                } catch (e) {
                    console.error("Failed to parse binary secret:", e);
                }
            }
        }
    }, [revealedSecret, revealedVersion, setValue, initialValues?.valueType, initialValues?.name]);

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

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

    async function handleDownloadFile() {
        if (!selectedFile) {
            return;
        }

        if (typeof window !== "undefined" && "showSaveFilePicker" in window) {
            try {
                const handle = await (
                    window as unknown as {
                        showSaveFilePicker: (options?: { suggestedName?: string }) => Promise<{
                            createWritable: () => Promise<{
                                write: (data: Blob) => Promise<void>;
                                close: () => Promise<void>;
                            }>;
                        }>;
                    }
                ).showSaveFilePicker({
                    suggestedName: selectedFile.name,
                });
                const writable = await handle.createWritable();
                await writable.write(selectedFile);
                await writable.close();
                toast.success("File saved successfully");
                return;
            } catch (err: unknown) {
                if (err && typeof err === "object" && "name" in err && err.name === "AbortError") {
                    return;
                }
            }
        }

        const url = window.URL.createObjectURL(selectedFile);
        const anchor = document.createElement("a");
        anchor.href = url;
        anchor.download = selectedFile.name;
        document.body.appendChild(anchor);
        anchor.click();
        document.body.removeChild(anchor);
        window.URL.revokeObjectURL(url);
    }

    function onValid(values: CreateOrEditProjectSecretFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditProjectSecretFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                if (readOnly) {
                    return;
                }

                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <fieldset
                disabled={readOnly}
                className="contents"
            >
                <div className="flex flex-col gap-6">
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
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                            <TabsList className="bg-muted/80 p-1 rounded-lg">
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
                                        placeholder={
                                            isEditMode ? "Leave empty to keep current value" : "Enter secret value"
                                        }
                                        rows={8}
                                        aria-invalid={isTextValueInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                                    <div className="flex flex-wrap items-center gap-3">
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
                                        {selectedFile ? (
                                            <div className="flex items-center gap-2">
                                                <span className="truncate text-sm font-medium text-foreground">
                                                    {selectedFile.name}
                                                </span>
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    className="h-8 gap-1.5 px-2.5 text-xs"
                                                    onClick={() => {
                                                        void handleDownloadFile();
                                                    }}
                                                >
                                                    <Download className="size-3.5" />
                                                    Download
                                                </Button>
                                            </div>
                                        ) : isEditMode ? (
                                            <span className="truncate text-sm text-muted-foreground">
                                                Leave empty to keep current value
                                            </span>
                                        ) : null}
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
                </div>
                {!readOnly && (
                    <FormActionBar sticky={stickyActions}>
                        <Button
                            type="button"
                            variant="outline"
                            className="min-w-[100px]"
                            disabled={isPending}
                            onClick={onClose}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            isLoading={isPending}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </FormActionBar>
                )}
                {readOnly && (
                    <FormActionBar sticky={stickyActions}>
                        <Button
                            type="button"
                            onClick={onClose}
                            className="min-w-[100px]"
                        >
                            Close
                        </Button>
                    </FormActionBar>
                )}
            </fieldset>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditProjectSecretFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    revealedSecret?: string | null;
    revealedVersion?: number;
    isEditMode: boolean;
    initialValues?: Partial<CreateOrEditProjectSecretFormInput>;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}
