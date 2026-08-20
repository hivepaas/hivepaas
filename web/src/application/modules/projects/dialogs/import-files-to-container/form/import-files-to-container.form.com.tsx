import { useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { UploadIcon } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Button, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import type { ImportFilesToContainerFormInput, ImportFilesToContainerFormOutput } from "../schemas";
import { ImportFilesToContainerFormSchema } from "../schemas";

const compressionOptions = [
    { value: "auto", label: "Auto" },
    { value: "tar", label: "Tar" },
    { value: "zip", label: "Zip" },
    { value: "gzip", label: "Gzip" },
    { value: "zstd", label: "Zstd" },
] as const;

export function ImportFilesToContainerForm({ isPending, onSubmit }: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<ImportFilesToContainerFormInput, unknown, ImportFilesToContainerFormOutput>({
        defaultValues: {
            file: undefined,
            path: "",
            extract: false,
            compression: "auto",
            overwrite: true,
        },
        resolver: zodResolver(ImportFilesToContainerFormSchema),
        mode: "onSubmit",
    });

    const { field: file } = useController({
        name: "file",
        control,
    });
    const { field: path } = useController({
        name: "path",
        control,
    });
    const { field: extract } = useController({
        name: "extract",
        control,
    });
    const { field: compression } = useController({
        name: "compression",
        control,
    });
    const { field: overwrite } = useController({
        name: "overwrite",
        control,
    });

    function onValid(values: ImportFilesToContainerFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<ImportFilesToContainerFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody>
                <FieldGroup>
                    <Field>
                        <InfoBlock
                            title="File"
                            titleWidth={130}
                        >
                            <div className="flex items-center gap-3">
                                <Button
                                    type="button"
                                    variant="outline"
                                    disabled={isPending}
                                    onClick={() => {
                                        fileInputRef.current?.click();
                                    }}
                                >
                                    <UploadIcon className="size-4" />
                                    Choose File
                                </Button>
                                <span className="truncate text-sm text-muted-foreground">
                                    {file.value instanceof File ? file.value.name : ""}
                                </span>
                            </div>
                            <Input
                                ref={fileInputRef}
                                type="file"
                                className="hidden"
                                disabled={isPending}
                                onChange={event => {
                                    file.onChange(event.target.files?.[0] ?? undefined);
                                }}
                            />
                            <FieldError errors={[errors.file]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock
                            title="Destination Path"
                            titleWidth={130}
                        >
                            <Input
                                {...path}
                                placeholder="/path/in/container"
                                disabled={isPending}
                            />
                            <FieldError errors={[errors.path]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock
                            title="Extract"
                            titleWidth={130}
                        >
                            <label
                                htmlFor="import-container-extract"
                                className="flex items-center gap-3 text-sm font-medium"
                            >
                                <Checkbox
                                    id="import-container-extract"
                                    checked={extract.value}
                                    disabled={isPending}
                                    onCheckedChange={checked => {
                                        extract.onChange(checked === true);
                                    }}
                                />
                            </label>
                        </InfoBlock>
                    </Field>

                    {extract.value && (
                        <Field>
                            <InfoBlock
                                title="Compression"
                                titleWidth={130}
                            >
                                <Tabs
                                    value={compression.value}
                                    onValueChange={compression.onChange}
                                >
                                    <TabsList>
                                        {compressionOptions.map(option => (
                                            <TabsTrigger
                                                key={option.value}
                                                value={option.value}
                                                disabled={isPending}
                                            >
                                                {option.label}
                                            </TabsTrigger>
                                        ))}
                                    </TabsList>
                                </Tabs>
                                <FieldError errors={[errors.compression]} />
                            </InfoBlock>
                        </Field>
                    )}

                    <Field>
                        <InfoBlock
                            title="Overwrite"
                            titleWidth={130}
                        >
                            <label
                                htmlFor="import-container-overwrite"
                                className="flex items-center gap-3 text-sm font-medium"
                            >
                                <Checkbox
                                    id="import-container-overwrite"
                                    checked={overwrite.value}
                                    disabled={isPending}
                                    onCheckedChange={checked => {
                                        overwrite.onChange(checked === true);
                                    }}
                                />
                            </label>
                        </InfoBlock>
                    </Field>
                </FieldGroup>
            </DialogBody>

            <DialogActionFooter className="flex justify-end gap-4">
                <Button
                    type="submit"
                    disabled={isPending}
                >
                    {isPending ? "Uploading..." : "Upload"}
                </Button>
            </DialogActionFooter>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: ImportFilesToContainerFormOutput) => void | Promise<void>;
}
