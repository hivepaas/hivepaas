import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Button, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

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
                        <InfoBlock
                            title="Path"
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
                            title="Is Directory"
                            titleWidth={130}
                        >
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
                        <InfoBlock
                            title="Compress"
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
