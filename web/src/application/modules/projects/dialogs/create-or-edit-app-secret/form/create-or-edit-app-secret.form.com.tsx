import React, { useEffect, useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { UploadIcon } from "lucide-react";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditAppSecretFormInput, CreateOrEditAppSecretFormOutput } from "../schemas";
import { APP_SECRET_DEFAULT_FILE_MODE, CreateOrEditAppSecretFormSchema } from "../schemas";

function getDefaultFilePath(name: string): string {
    const normalizedName = name.trim().toLowerCase();
    return normalizedName ? `/run/secrets/${normalizedName}` : "/run/secrets/secret_name";
}

export function CreateOrEditAppSecretForm({
    isPending,
    onSubmit,
    onHasChanges,
    isEditMode,
    initialValues,
    readOnly = false,
    onClose,
}: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    const {
        handleSubmit,
        control,
        setValue,
        formState: { errors, isDirty, dirtyFields },
    } = useForm<CreateOrEditAppSecretFormInput, unknown, CreateOrEditAppSecretFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            valueType: initialValues?.valueType ?? "text",
            isEditMode,
            textValue: "",
            binaryFile: null,
            mountIntoFilesystem: initialValues?.mountIntoFilesystem ?? false,
            filePath: initialValues?.filePath ?? getDefaultFilePath(initialValues?.name ?? ""),
            fileMode: initialValues?.fileMode ?? APP_SECRET_DEFAULT_FILE_MODE,
            fileUid: initialValues?.fileUid ?? "",
            fileGid: initialValues?.fileGid ?? "",
        },
        resolver: zodResolver(CreateOrEditAppSecretFormSchema),
        mode: "onSubmit",
    });

    const secretName = useWatch({ control, name: "name" });
    const valueType = useWatch({ control, name: "valueType" });
    const mountIntoFilesystem = useWatch({ control, name: "mountIntoFilesystem" });
    const selectedFile = useWatch({ control, name: "binaryFile" });

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    useEffect(() => {
        if (isEditMode || dirtyFields.filePath) {
            return;
        }

        setValue("filePath", getDefaultFilePath(secretName), { shouldDirty: false });
    }, [dirtyFields.filePath, isEditMode, secretName, setValue]);

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

    const { field: mountIntoFilesystemField } = useController({
        name: "mountIntoFilesystem",
        control,
    });

    const { field: binaryFileField } = useController({
        name: "binaryFile",
        control,
    });

    const {
        field: filePath,
        fieldState: { invalid: isFilePathInvalid },
    } = useController({
        name: "filePath",
        control,
    });

    const {
        field: fileMode,
        fieldState: { invalid: isFileModeInvalid },
    } = useController({
        name: "fileMode",
        control,
    });

    const { field: fileUid } = useController({
        name: "fileUid",
        control,
    });

    const { field: fileGid } = useController({
        name: "fileGid",
        control,
    });

    function onValid(values: CreateOrEditAppSecretFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditAppSecretFormOutput>) {
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
                                    id="app-secret-name"
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
                                        id="app-secret-text-value"
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
                                            {selectedFile?.name ??
                                                (isEditMode ? "Leave empty to keep current value" : "")}
                                        </span>
                                    </div>
                                    <Input
                                        id="app-secret-binary-value"
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

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Mount into Filesystem" />}
                    >
                        <Checkbox
                            checked={mountIntoFilesystem}
                            onCheckedChange={checked => {
                                mountIntoFilesystemField.onChange(checked === true);
                            }}
                        />
                    </InfoBlock>

                    {mountIntoFilesystem && (
                        <>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="File Path"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            id="app-secret-file-path"
                                            {...filePath}
                                            placeholder="/run/secrets/secret_name"
                                            aria-invalid={isFilePathInvalid}
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                        />
                                        <FieldError errors={[errors.filePath]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="File Mode"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            id="app-secret-file-mode"
                                            {...fileMode}
                                            placeholder="default: 0444"
                                            aria-invalid={isFileModeInvalid}
                                            className="max-w-[180px]"
                                        />
                                        <FieldError errors={[errors.fileMode]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={<LabelWithInfo label="File UID" />}
                            >
                                <Input
                                    id="app-secret-file-uid"
                                    {...fileUid}
                                    placeholder="uid"
                                    className="max-w-[180px]"
                                />
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={220}
                                title={<LabelWithInfo label="File GID" />}
                            >
                                <Input
                                    id="app-secret-file-gid"
                                    {...fileGid}
                                    placeholder="gid"
                                    className="max-w-[180px]"
                                />
                            </InfoBlock>
                        </>
                    )}
                </div>
                {!readOnly && (
                    <div className="pb-6 flex justify-end mt-6">
                        <div className="flex items-center gap-3">
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
                        </div>
                    </div>
                )}
                {readOnly && (
                    <div className="shrink-0 px-0 mt-6 pb-6 flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                        >
                            Close
                        </Button>
                    </div>
                )}
            </fieldset>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditAppSecretFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    isEditMode: boolean;
    initialValues?: Partial<CreateOrEditAppSecretFormInput>;
    readOnly?: boolean;
    onClose?: () => void;
}
