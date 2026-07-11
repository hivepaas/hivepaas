import React, { useMemo, useRef, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { UploadIcon } from "lucide-react";
import { useController, useForm, useWatch } from "react-hook-form";
import { ProjectCloudStorageQueries } from "~/projects/data/queries";
import { APP_DATA_FILE_KIND_OPTIONS, PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import {
    AppLink,
    Combobox,
    EditableCombobox,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import { Button, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import type { UploadAppDataFileFormValues } from "../schemas";
import { UploadAppDataFileFormSchema } from "../schemas";

type StorageOption = { id: string; name: string };

export function UploadAppDataFileForm({ isPending, onSubmit, projectId, onClose }: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [storageSearch, setStorageSearch] = useState("");

    const {
        data: { data: cloudStorages } = DEFAULT_PAGINATED_DATA,
        isFetching: isStorageFetching,
        refetch: refetchStorage,
        isRefetching: isStorageRefetching,
    } = ProjectCloudStorageQueries.useFindManyPaginated({ projectID: projectId, search: storageSearch });

    const storageOptions = useMemo(
        () =>
            cloudStorages.map(item => ({
                value: { id: item.id, name: item.name } satisfies StorageOption,
                label: item.name,
            })),
        [cloudStorages],
    );

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<UploadAppDataFileFormValues>({
        defaultValues: {
            uploadMethod: "local",
            files: [],
            fileKind: "",
        },
        resolver: zodResolver(UploadAppDataFileFormSchema),
        mode: "onSubmit",
    });

    const uploadMethod = useWatch({ control, name: "uploadMethod" as never }) as unknown as "local" | "cloud";
    const selectedFiles = useWatch({ control, name: "files" as never }) as File[] | undefined;

    const { field: uploadMethodField } = useController({ name: "uploadMethod", control });

    const {
        field: filesField,
        fieldState: { invalid: isFilesInvalid },
    } = useController({
        name: "files" as never,
        control,
    });

    const {
        field: fileKindLocalField,
        fieldState: { invalid: isFileKindLocalInvalid },
    } = useController({
        name: "fileKind" as never,
        control,
    });

    const {
        field: storageField,
        fieldState: { invalid: isStorageInvalid },
    } = useController({
        name: "storage" as never,
        control,
    });

    const {
        field: bucketField,
        fieldState: { invalid: isBucketInvalid },
    } = useController({
        name: "bucket" as never,
        control,
    });

    const {
        field: fileKindCloudField,
        fieldState: { invalid: isFileKindCloudInvalid },
    } = useController({
        name: "fileKind" as never,
        control,
    });

    const {
        field: filePathField,
        fieldState: { invalid: isFilePathInvalid },
    } = useController({
        name: "filePath" as never,
        control,
    });

    const localErrors = errors as Record<string, { message?: string } | undefined>;

    const fileLabel =
        Array.isArray(selectedFiles) && selectedFiles.length > 0
            ? selectedFiles.length === 1
                ? ((selectedFiles[0] as File | undefined)?.name ?? "")
                : `${selectedFiles.length} files selected`
            : "";

    const cloudStoragesRoute = ROUTE.projects.single.providerConfiguration.cloudStorages.$route(projectId);

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(values => onSubmit(values))(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <div className="flex flex-col gap-6">
                <InfoBlock
                    titleWidth={200}
                    title={
                        <LabelWithInfo
                            label="Upload Method"
                            isRequired
                        />
                    }
                >
                    <Tabs
                        value={uploadMethod}
                        onValueChange={value => {
                            uploadMethodField.onChange(value);
                        }}
                    >
                        <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                            <TabsTrigger value="local">Local Files</TabsTrigger>
                            <TabsTrigger value="cloud">File in Cloud</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {uploadMethod === "local" && (
                    <>
                        <InfoBlock
                            titleWidth={200}
                            title={
                                <LabelWithInfo
                                    label="Files"
                                    isRequired
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
                                            Choose Files
                                        </Button>
                                        {fileLabel && (
                                            <span className="truncate text-sm text-muted-foreground">{fileLabel}</span>
                                        )}
                                    </div>
                                    <Input
                                        ref={fileInputRef}
                                        type="file"
                                        multiple
                                        className="hidden"
                                        aria-invalid={isFilesInvalid}
                                        onChange={event => {
                                            const fileList = event.target.files;
                                            filesField.onChange(fileList ? Array.from(fileList) : []);
                                        }}
                                    />
                                    <FieldError errors={[localErrors["files"]]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={200}
                            title={
                                <LabelWithInfo
                                    label="File Kind"
                                    isRequired
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <EditableCombobox
                                        options={APP_DATA_FILE_KIND_OPTIONS}
                                        value={fileKindLocalField.value as string}
                                        onChange={fileKindLocalField.onChange}
                                        placeholder="select or type file kind"
                                        aria-invalid={isFileKindLocalInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[localErrors["fileKind"]]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </>
                )}

                {uploadMethod === "cloud" && (
                    <>
                        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                            <span className="text-orange-500">Note:</span> Uploading a file from cloud storage (S3) only
                            creates a database record referencing the file; the file content is not downloaded to the
                            local system. If you delete the file from the cloud storage, local file operations may no
                            longer work.
                        </div>

                        <InfoBlock
                            titleWidth={200}
                            title={
                                <LabelWithInfo
                                    label="Storage"
                                    isRequired
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <Combobox<StorageOption>
                                        options={storageOptions}
                                        value={(storageField.value as StorageOption | null | undefined)?.id ?? null}
                                        onChange={(_, option) => {
                                            storageField.onChange(option ?? null);
                                        }}
                                        onSearch={setStorageSearch}
                                        placeholder="select cloud storage"
                                        searchable
                                        allowClear
                                        closeOnSelect
                                        emptyText="No cloud storages available"
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                        valueKey="id"
                                        aria-invalid={isStorageInvalid}
                                        loading={isStorageFetching}
                                        onRefresh={() => void refetchStorage()}
                                        isRefreshing={isStorageRefetching}
                                    />
                                    <FieldError errors={[localErrors["storage"]]} />
                                    <AppLink.Modules
                                        to={cloudStoragesRoute}
                                        className="text-sm text-link"
                                        ignorePrevPath
                                    >
                                        Configure storage settings
                                    </AppLink.Modules>
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={200}
                            title={
                                <LabelWithInfo
                                    label="Bucket"
                                    isRequired
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...bucketField}
                                        placeholder="bucket name"
                                        aria-invalid={isBucketInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[localErrors["bucket"]]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={200}
                            title={
                                <LabelWithInfo
                                    label="File Kind"
                                    isRequired
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <EditableCombobox
                                        options={APP_DATA_FILE_KIND_OPTIONS}
                                        value={fileKindCloudField.value as string}
                                        onChange={fileKindCloudField.onChange}
                                        placeholder="select or type file kind"
                                        aria-invalid={isFileKindCloudInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[localErrors["fileKind"]]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={200}
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
                                        {...filePathField}
                                        placeholder="full path to file"
                                        aria-invalid={isFilePathInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[localErrors["filePath"]]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </>
                )}
            </div>

            <FormActionBar sticky>
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
                    Upload
                </Button>
            </FormActionBar>
        </form>
    );
}

interface Props {
    isPending: boolean;
    projectId: string;
    onSubmit: (values: UploadAppDataFileFormValues) => Promise<void> | void;
    onClose?: () => void;
}
