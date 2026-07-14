import { useState } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { useController, useFormContext } from "react-hook-form";
import { ProjectCloudStorageQueries } from "~/projects/data/queries";

import { AppLink, Combobox, ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import type { CreateOrEditAppScheduledJobFormInput } from "../../schemas";

type SchemaInput = CreateOrEditAppScheduledJobFormInput;

const COMPRESSION_OPTIONS = [
    { value: "", label: "Disabled" },
    { value: "gzip", label: "Gzip" },
    { value: "zstd", label: "Zstd" },
] as const;

const ENCRYPTION_OPTIONS = [
    { value: "", label: "Disabled" },
    { value: "age", label: "Age" },
] as const;

interface StorageOption {
    [key: string]: unknown;
    id: string;
    name: string;
}

interface Props {
    projectId: string;
    readOnly?: boolean;
}

export function SaveToFileSection({ projectId, readOnly = false }: Props) {
    const [storageSearch, setStorageSearch] = useState("");

    const { control } = useFormContext<SchemaInput>();

    const {
        field: fileName,
        fieldState: { invalid: isFileNameInvalid },
    } = useController({ control, name: "saveToFile.fileName" });
    const { field: compressionFormat } = useController({ control, name: "saveToFile.compressionFormat" });
    const { field: encryptionFormat } = useController({ control, name: "saveToFile.encryptionFormat" });
    const {
        field: encryptionSecret,
        fieldState: { invalid: isEncryptionSecretInvalid, error: encryptionSecretError },
    } = useController({ control, name: "saveToFile.encryptionSecret" });
    const {
        field: storage,
        fieldState: { invalid: isStorageInvalid },
    } = useController({ control, name: "saveToFile.storage" });
    const { field: filePath } = useController({ control, name: "saveToFile.filePath" });

    const {
        data: storageData,
        isFetching: isStorageFetching,
        isRefetching: isStorageRefetching,
        refetch: refetchStorage,
    } = ProjectCloudStorageQueries.useFindManyPaginated({
        projectID: projectId,
        search: storageSearch,
    });

    const storageOptions = (storageData?.data ?? []).map(s => ({
        value: { id: s.id, name: s.name } satisfies StorageOption,
        label: s.name,
    }));

    const cloudStoragesRoute = ROUTE.projects.single.providerConfiguration.cloudStorages.$route(projectId);

    return (
        <ContentBlock label="Command Output: Save to File">
            <div className="flex flex-col gap-6">
                <InfoBlock
                    title={<LabelWithInfo label="File Name" />}
                    titleWidth={160}
                >
                    <Field>
                        <Input
                            {...fileName}
                            placeholder="my_backup_%Y-%m-%d_%H:%M:%S.sql"
                            className="max-w-[500px]"
                            aria-invalid={isFileNameInvalid}
                            disabled={readOnly}
                        />
                        <FieldError errors={[]} />
                    </Field>
                </InfoBlock>

                <InfoBlock
                    title="Compress Data"
                    titleWidth={160}
                >
                    <Tabs
                        value={compressionFormat.value}
                        onValueChange={compressionFormat.onChange}
                    >
                        <TabsList>
                            {COMPRESSION_OPTIONS.map(opt => (
                                <TabsTrigger
                                    key={opt.value}
                                    value={opt.value}
                                    disabled={readOnly}
                                >
                                    {opt.label}
                                </TabsTrigger>
                            ))}
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                <InfoBlock
                    title="Encrypt Data"
                    titleWidth={160}
                >
                    <Tabs
                        value={encryptionFormat.value}
                        onValueChange={encryptionFormat.onChange}
                    >
                        <TabsList>
                            {ENCRYPTION_OPTIONS.map(opt => (
                                <TabsTrigger
                                    key={opt.value}
                                    value={opt.value}
                                    disabled={readOnly}
                                >
                                    {opt.label}
                                </TabsTrigger>
                            ))}
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {encryptionFormat.value !== "" && (
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Encryption Secret"
                                isRequired
                            />
                        }
                        titleWidth={160}
                    >
                        <FieldGroup>
                            <Field className="max-w-[400px]">
                                <PasswordInput
                                    value={encryptionSecret.value}
                                    onChange={encryptionSecret.onChange}
                                    onBlur={encryptionSecret.onBlur}
                                    placeholder="password"
                                    className="max-w-[400px]"
                                    aria-invalid={isEncryptionSecretInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[encryptionSecretError]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                )}

                <InfoBlock
                    title="Save File in Cloud Storage"
                    titleWidth={160}
                >
                    <FieldGroup>
                        <Field>
                            <Combobox<StorageOption>
                                options={storageOptions}
                                value={(storage.value as StorageOption | null)?.id ?? null}
                                onChange={(_, option) => {
                                    storage.onChange(option ?? null);
                                }}
                                onSearch={setStorageSearch}
                                placeholder="Select cloud storage"
                                emptyText="No cloud storages available"
                                className="max-w-[400px]"
                                valueKey="id"
                                searchable
                                closeOnSelect
                                allowClear
                                loading={isStorageFetching}
                                onRefresh={() => void refetchStorage()}
                                isRefreshing={isStorageRefetching}
                                aria-invalid={isStorageInvalid}
                                disabled={readOnly}
                            />
                            <AppLink.Modules
                                to={cloudStoragesRoute}
                                className="text-sm text-primary hover:underline mt-1 block"
                            >
                                Configure Cloud Storages
                            </AppLink.Modules>
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                {storage.value && (
                    <InfoBlock
                        title="Destination Directory"
                        titleWidth={160}
                    >
                        <Field>
                            <Input
                                {...filePath}
                                placeholder="path/to/sub/dir"
                                className="max-w-[400px]"
                                disabled={readOnly}
                            />
                        </Field>
                    </InfoBlock>
                )}
            </div>
        </ContentBlock>
    );
}
