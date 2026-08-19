import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

import type { ExportContainerFilesFormInput, ExportContainerFilesFormOutput } from "../schemas";
import { ExportContainerFilesFormSchema } from "../schemas";

const compressionOptions = [
    { value: "none", label: "None" },
    { value: "gzip", label: "Gzip" },
    { value: "zstd", label: "Zstd" },
] as const;

export function ExportContainerFilesForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<ExportContainerFilesFormInput, unknown, ExportContainerFilesFormOutput>({
        defaultValues: {
            path: "",
            isDir: false,
            compression: "none",
        },
        resolver: zodResolver(ExportContainerFilesFormSchema),
        mode: "onSubmit",
    });

    const { field: path } = useController({
        name: "path",
        control,
    });
    const { field: isDir } = useController({
        name: "isDir",
        control,
    });
    const { field: compression } = useController({
        name: "compression",
        control,
    });

    function onValid(values: ExportContainerFilesFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<ExportContainerFilesFormOutput>) {
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
                        <InfoBlock title="Path">
                            <Input
                                {...path}
                                placeholder="/path/in/container"
                                disabled={isPending}
                            />
                            <FieldError errors={[errors.path]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock title="Is Directory">
                            <label
                                htmlFor="export-container-is-dir"
                                className="flex items-center gap-3 text-sm font-medium"
                            >
                                <Checkbox
                                    id="export-container-is-dir"
                                    checked={isDir.value}
                                    disabled={isPending}
                                    onCheckedChange={checked => {
                                        isDir.onChange(checked === true);
                                    }}
                                />
                            </label>
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock title="Compress">
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
                </FieldGroup>
            </DialogBody>

            <DialogActionFooter className="flex justify-end gap-4">
                <Button
                    type="submit"
                    disabled={isPending}
                >
                    {isPending ? "Downloading..." : "Download"}
                </Button>
            </DialogActionFooter>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: ExportContainerFilesFormOutput) => void | Promise<void>;
}
