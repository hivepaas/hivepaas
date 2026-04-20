import { useState } from "react";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, Plus, Trash2, X } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputNumberWithAddon } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface CompressionConfigSectionProps {
    prefix: string;
    onRemove?: () => void;
}

export function CompressionConfigSection({ prefix, onRemove }: CompressionConfigSectionProps) {
    const [open, setOpen] = useState(false);
    const [excludedInput, setExcludedInput] = useState("");
    const [includedInput, setIncludedInput] = useState("");

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const { field: enabled } = useController({ control, name: `${prefix}.enabled` as never });
    const {
        field: defaultEncoding,
        fieldState: { error: defaultEncodingError },
    } = useController({ control, name: `${prefix}.defaultEncoding` as never });
    const {
        field: minResponseBody,
        fieldState: { error: minResponseBodyError },
    } = useController({ control, name: `${prefix}.minResponseBody` as never });

    const {
        fields: excludedFields,
        append: appendExcluded,
        remove: removeExcluded,
    } = useFieldArray({ control, name: `${prefix}.excludedContentTypes` as never });

    const {
        fields: includedFields,
        append: appendIncluded,
        remove: removeIncluded,
    } = useFieldArray({ control, name: `${prefix}.includedContentTypes` as never });

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <div className="flex w-full items-center gap-1 rounded-md border border-dashed px-1">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex min-w-0 flex-1 items-center gap-2 px-2 py-2 text-sm font-medium hover:bg-accent"
                    >
                        {open ? (
                            <ChevronDown className="size-4 shrink-0" />
                        ) : (
                            <ChevronRight className="size-4 shrink-0" />
                        )}
                        Compression Configuration
                    </button>
                </CollapsibleTrigger>
                {onRemove && (
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="shrink-0 text-muted-foreground hover:text-destructive"
                        title="Remove section"
                        onClick={onRemove}
                    >
                        <X className="size-4" />
                    </Button>
                )}
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Enabled">
                        <Checkbox
                            checked={enabled.value}
                            onCheckedChange={enabled.onChange}
                        />
                    </InfoBlock>

                    <InfoBlock title="Default Encoding">
                        <Input
                            {...defaultEncoding}
                            value={defaultEncoding.value}
                            onChange={defaultEncoding.onChange}
                            placeholder="gzip"
                            className="max-w-[300px]"
                        />
                        <FieldError errors={[defaultEncodingError]} />
                    </InfoBlock>

                    <InfoBlock title="Min Response Body Size">
                        <InputNumberWithAddon
                            addonLeft="Bytes"
                            value={minResponseBody.value}
                            onValueChange={v => {
                                minResponseBody.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[300px]"
                        />
                        <FieldError errors={[minResponseBodyError]} />
                    </InfoBlock>

                    <InfoBlock title="Excluded Content Types">
                        <div className="flex flex-col gap-2 max-w-[400px]">
                            <div className="flex gap-2">
                                <Input
                                    value={excludedInput}
                                    onChange={e => {
                                        setExcludedInput(e.target.value);
                                    }}
                                    placeholder="text/plain"
                                    className="flex-1"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        if (excludedInput.trim()) {
                                            appendExcluded({ value: excludedInput.trim() });
                                            setExcludedInput("");
                                        }
                                    }}
                                >
                                    <Plus className="size-4" />
                                </Button>
                            </div>
                            <div className="flex flex-col divide-y">
                                {excludedFields.map((field, index) => (
                                    <div
                                        key={field.id}
                                        className="flex items-center justify-between py-1.5"
                                    >
                                        <span className="text-sm">{(field as unknown as { value: string }).value}</span>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => {
                                                removeExcluded(index);
                                            }}
                                        >
                                            <Trash2 className="size-3.5" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </InfoBlock>

                    <InfoBlock title="Included Content Types">
                        <div className="flex flex-col gap-2 max-w-[400px]">
                            <div className="flex gap-2">
                                <Input
                                    value={includedInput}
                                    onChange={e => {
                                        setIncludedInput(e.target.value);
                                    }}
                                    placeholder="text/html"
                                    className="flex-1"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        if (includedInput.trim()) {
                                            appendIncluded({ value: includedInput.trim() });
                                            setIncludedInput("");
                                        }
                                    }}
                                >
                                    <Plus className="size-4" />
                                </Button>
                            </div>
                            <div className="flex flex-col divide-y">
                                {includedFields.map((field, index) => (
                                    <div
                                        key={field.id}
                                        className="flex items-center justify-between py-1.5"
                                    >
                                        <span className="text-sm">{(field as unknown as { value: string }).value}</span>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => {
                                                removeIncluded(index);
                                            }}
                                        >
                                            <Trash2 className="size-3.5" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
