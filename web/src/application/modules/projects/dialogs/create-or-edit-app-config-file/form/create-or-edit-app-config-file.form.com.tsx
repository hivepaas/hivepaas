import React, { useEffect, useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { DownloadIcon, UploadIcon } from "lucide-react";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { AvailableInAppsWarning, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditAppConfigFileFormInput, CreateOrEditAppConfigFileFormOutput } from "../schemas";
import { CreateOrEditAppConfigFileFormSchema } from "../schemas";

export function CreateOrEditAppConfigFileForm({
    isPending,
    onSubmit,
    onHasChanges,
    isEditMode,
    initialValues,
    readOnly = false,
    stickyActions = false,
    onClose,
    inheritableLabel = "Available in Previews",
    inheritableWarning = "Warning: Preview apps will not be able to access this configuration.",
    currentBinaryContent,
}: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditAppConfigFileFormInput, unknown, CreateOrEditAppConfigFileFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            valueType: initialValues?.valueType ?? "text",
            isEditMode,
            initialValueType: isEditMode ? (initialValues?.valueType ?? null) : null,
            // A text config file's content is shown and edited in place; a
            // binary one is replaced by an upload.
            textValue: initialValues?.textValue ?? "",
            binaryFile: null,
            // New settings are available unless the person says otherwise.
            inheritable: initialValues?.inheritable ?? true,
        },
        resolver: zodResolver(CreateOrEditAppConfigFileFormSchema),
        mode: "onSubmit",
    });

    const valueType = useWatch({ control, name: "valueType" });
    const keepsBinary = isEditMode && initialValues?.valueType === "binary";
    const selectedFile = useWatch({ control, name: "binaryFile" });

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

    const { field: inheritableField } = useController({
        name: "inheritable",
        control,
    });

    const { field: binaryFileField } = useController({
        name: "binaryFile",
        control,
    });

    function onValid(values: CreateOrEditAppConfigFileFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditAppConfigFileFormOutput>) {
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
                                    id="app-config-file-name"
                                    {...name}
                                    placeholder="CONFIG_NAME"
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
                                    isRequired
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <Textarea
                                        id="app-config-file-text-value"
                                        {...textValue}
                                        placeholder="Enter config content"
                                        rows={8}
                                        aria-invalid={isTextValueInvalid}
                                        className="w-full"
                                    />
                                    <p className="text-sm text-muted-foreground">Max size: 1mb</p>
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
                                    isRequired={!keepsBinary}
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
                                        {keepsBinary && currentBinaryContent && (
                                            // A link, not a button: a read-only form's disabled
                                            // fieldset disables its buttons, and the file is still
                                            // worth saving there.
                                            <Button
                                                asChild
                                                variant="outline"
                                            >
                                                <a
                                                    href={`data:application/octet-stream;base64,${currentBinaryContent}`}
                                                    download={name.value}
                                                >
                                                    <DownloadIcon className="size-4" />
                                                    Download Current File
                                                </a>
                                            </Button>
                                        )}
                                        <span className="truncate text-sm text-muted-foreground">
                                            {selectedFile?.name ??
                                                (keepsBinary ? "Leave empty to keep current content" : "")}
                                        </span>
                                    </div>
                                    <Input
                                        id="app-config-file-binary-value"
                                        ref={fileInputRef}
                                        type="file"
                                        className="hidden"
                                        onChange={event => {
                                            binaryFileField.onChange(event.target.files?.[0] ?? null);
                                        }}
                                    />
                                    <p className="text-sm text-muted-foreground">Max size: 1mb</p>
                                    <FieldError errors={[errors.binaryFile]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    )}

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label={inheritableLabel} />}
                    >
                        <div className="flex items-center gap-3">
                            <Checkbox
                                id="app-config-file-inheritable"
                                checked={inheritableField.value}
                                onCheckedChange={checked => {
                                    inheritableField.onChange(Boolean(checked));
                                }}
                            />
                            {!inheritableField.value ? <AvailableInAppsWarning message={inheritableWarning} /> : null}
                        </div>
                    </InfoBlock>
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
    onSubmit: (values: CreateOrEditAppConfigFileFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    isEditMode: boolean;
    initialValues?: Partial<CreateOrEditAppConfigFileFormInput>;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
    /** An app's config file is available to its previews; a project's or env's to its apps. */
    inheritableLabel?: string;
    inheritableWarning?: string;
    /** A binary config file's content in base64, offered for download while editing it. */
    currentBinaryContent?: string;
}
