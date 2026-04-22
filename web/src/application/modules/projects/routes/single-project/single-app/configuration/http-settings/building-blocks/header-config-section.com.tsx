import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, Plus, Trash2, X } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput } from "../schemas";

interface HeaderConfigSectionProps {
    prefix: string;
    onRemove?: () => void;
}

function KeyValueList({
    name,
    title,
    keyPlaceholder,
    valuePlaceholder,
    description,
}: {
    name: string;
    title: string;
    keyPlaceholder?: string;
    valuePlaceholder?: string;
    description?: string;
}) {
    const { control } = useFormContext<AppConfigHttpSettingsFormSchemaInput>();
    const { fields, append, remove } = useFieldArray({ control, name: name as never });
    const [keyInput, setKeyInput] = useState("");
    const [valueInput, setValueInput] = useState("");

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label={title}
                    content={description}
                />
            }
        >
            <div className="flex flex-col gap-2 max-w-[550px]">
                <div className="flex gap-2">
                    <div className="flex-1 grid grid-cols-2 gap-2">
                        <InputWithAddOn
                            addonLeft="Name"
                            value={keyInput}
                            onChange={e => {
                                setKeyInput(e.target.value);
                            }}
                            placeholder={keyPlaceholder ?? "Name"}
                            className="flex-1"
                        />
                        <InputWithAddOn
                            addonLeft="Value"
                            value={valueInput}
                            onChange={e => {
                                setValueInput(e.target.value);
                            }}
                            placeholder={valuePlaceholder ?? "Value"}
                            className="flex-1"
                        />
                    </div>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            if (keyInput.trim()) {
                                append({ key: keyInput.trim(), value: valueInput.trim() });
                                setKeyInput("");
                                setValueInput("");
                            }
                        }}
                    >
                        <Plus className="size-4" /> Add
                    </Button>
                </div>
                <div className="flex flex-col divide-y">
                    {fields.map((field, index) => (
                        <div
                            key={field.id}
                            className="flex items-center justify-between"
                        >
                            <div className="flex-1 grid grid-cols-2 items-center gap-2 py-1.5">
                                <span className="text-sm break-words">{(field as { key?: string }).key}</span>
                                <span className="text-sm break-words">{(field as { value?: string }).value}</span>
                            </div>
                            <div className="w-[76px]">
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => {
                                        remove(index);
                                    }}
                                >
                                    <Trash2 className="size-3.5" />
                                </Button>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </InfoBlock>
    );
}

function StringList({
    name,
    title,
    placeholder,
    description,
}: {
    name: string;
    title: string;
    placeholder?: string;
    description?: string;
}) {
    const { control } = useFormContext<AppConfigHttpSettingsFormSchemaInput>();
    const { fields, append, remove } = useFieldArray({ control, name: name as never });
    const [input, setInput] = useState("");

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label={title}
                    content={description}
                />
            }
        >
            <div className="flex flex-col gap-2 max-w-[314px]">
                <div className="flex gap-2">
                    <Input
                        value={input}
                        onChange={e => {
                            setInput(e.target.value);
                        }}
                        placeholder={placeholder ?? "Header name"}
                        className="flex-1"
                    />
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            if (input.trim()) {
                                append({ value: input.trim() });
                                setInput("");
                            }
                        }}
                    >
                        <Plus className="size-4" /> Add
                    </Button>
                </div>
                <div className="flex flex-col divide-y">
                    {fields.map((field, index) => (
                        <div
                            key={field.id}
                            className="flex items-center justify-between py-1.5"
                        >
                            <div className="flex-1 grid grid-cols-1 items-center">
                                <div className="text-sm break-words">{(field as { value?: string }).value}</div>
                            </div>
                            <div className="w-[76px]">
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => {
                                        remove(index);
                                    }}
                                >
                                    <Trash2 className="size-3.5" />
                                </Button>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </InfoBlock>
    );
}

export function HeaderConfigSection({ prefix, onRemove }: HeaderConfigSectionProps) {
    const [open, setOpen] = useState(false);

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
                        Header Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/headers/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            (docs)
                        </a>
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
                    <KeyValueList
                        name={`${prefix}.toAddToRequests`}
                        title="To Add To Requests"
                        description="Headers to add to the request."
                    />
                    <StringList
                        name={`${prefix}.toRemoveFromRequests`}
                        title="To Remove From Requests"
                        description="Headers to remove from the request."
                    />
                    <KeyValueList
                        name={`${prefix}.toAddToResponses`}
                        title="To Add To Responses"
                        description="Headers to add to the response."
                    />
                    <StringList
                        name={`${prefix}.toRemoveFromResponses`}
                        title="To Remove From Responses"
                        description="Headers to remove from the response."
                    />
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
