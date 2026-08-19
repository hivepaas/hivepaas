import { useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { UploadIcon } from "lucide-react";

import { InfoBlock } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

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
                        <InfoBlock title="File">
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
                        <InfoBlock title="Destination Path">
                            <Input
                                {...path}
                                placeholder="/path/in/container"
                                disabled={isPending}
                            />
                            <FieldError errors={[errors.path]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock title="Extract">
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
                            <InfoBlock title="Compression">
                                <div className="flex flex-wrap gap-2">
                                    {compressionOptions.map(option => (
                                        <Button
                                            key={option.value}
                                            type="button"
                                            variant="outline"
                                            disabled={isPending}
                                            className={cn(
                                                "min-w-[72px]",
                                                compression.value === option.value &&
                                                    "border-primary bg-background shadow-xs",
                                            )}
                                            onClick={() => {
                                                compression.onChange(option.value);
                                            }}
                                        >
                                            {option.label}
                                        </Button>
                                    ))}
                                </div>
                                <FieldError errors={[errors.compression]} />
                            </InfoBlock>
                        </Field>
                    )}

                    <Field>
                        <InfoBlock title="Overwrite">
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
