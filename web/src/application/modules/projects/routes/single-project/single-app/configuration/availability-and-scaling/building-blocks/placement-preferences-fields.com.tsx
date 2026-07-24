import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Check, Pencil, Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InfoBlock, LabelWithInfo, PopConfirm } from "@application/shared/components";

import { type AppConfigAvailabilitySchemaInput, type AppConfigAvailabilitySchemaOutput } from "../schemas";

export function PlacementPreferencesFields() {
    const { control } = useFormContext<AppConfigAvailabilitySchemaInput, unknown, AppConfigAvailabilitySchemaOutput>();

    const { fields, append, remove, update } = useFieldArray({
        control,
        name: "preferences",
    });

    const [newValue, setNewValue] = useState("");
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftValue, setDraftValue] = useState("");
    const strategy = "spread";

    const handleAddPreference = () => {
        if (newValue) {
            append({ name: strategy, value: newValue });
            setNewValue("");
            setEditingIndex(null);
            setDraftValue("");
        } else {
            toast.error("Please provide a value for the placement preference");
        }
    };

    const handleStartEdit = (index: number, currentValue: string) => {
        setEditingIndex(index);
        setDraftValue(currentValue);
    };

    const handleSaveEdit = (index: number) => {
        if (editingIndex !== index) {
            return;
        }

        if (!draftValue.trim()) {
            toast.error("Please provide a value for the placement preference");
            return;
        }

        const currentPreference = fields[index];
        if (!currentPreference) {
            return;
        }

        update(index, {
            name: currentPreference.name,
            value: draftValue.trim(),
        });
        setEditingIndex(null);
        setDraftValue("");
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftValue("");
    };

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Placement Preferences"
                        content="Preferences that guide the scheduling of the service."
                    />
                }
            >
                <div className="flex flex-col gap-4 max-w-[800px]">
                    <div className="flex gap-4 flex-wrap items-center">
                        <div className="grid grid-cols-2 gap-3 flex-1">
                            <div className="flex items-center rounded-md shadow-xs bg-background border border-input">
                                <span className="px-3 text-sm border-r border-input bg-muted/50 rounded-l-md h-9 flex items-center">
                                    Strategy
                                </span>
                                <Select
                                    value={strategy}
                                    disabled
                                >
                                    <SelectTrigger className="flex-1 border-0 shadow-none focus:ring-0 rounded-l-none h-9 opacity-100 bg-muted/20">
                                        <SelectValue placeholder="Select strategy" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="spread">spread</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            <div className="flex items-center rounded-md shadow-xs bg-background border border-input flex-1">
                                <span className="px-3 text-sm border-r border-input bg-muted/50 rounded-l-md h-9 flex items-center">
                                    Value
                                </span>
                                <Input
                                    value={newValue}
                                    onChange={e => {
                                        setNewValue(e.target.value);
                                    }}
                                    className="border-0 shadow-none focus-visible:ring-0 rounded-l-none h-9"
                                    placeholder="value"
                                />
                            </div>
                        </div>

                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAddPreference}
                            disabled={!newValue}
                            className="h-9 px-4 shrink-0"
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>

                    <div className="mt-2 divide-y divide-zinc-200">
                        {fields.map((field, index) => {
                            const isEditing = editingIndex === index;

                            return (
                                <div
                                    key={field.id}
                                    className="flex items-center group gap-4 py-2"
                                >
                                    <div className="grid grid-cols-2 flex-1 gap-3">
                                        <div className="text-sm wrap-break-word pl-3">{field.name}</div>
                                        {isEditing ? (
                                            <Input
                                                value={draftValue}
                                                onChange={e => {
                                                    setDraftValue(e.target.value);
                                                }}
                                                onKeyDown={e => {
                                                    if (e.key === "Enter") {
                                                        e.preventDefault();
                                                        handleSaveEdit(index);
                                                    }
                                                    if (e.key === "Escape") {
                                                        e.preventDefault();
                                                        handleCancelEdit();
                                                    }
                                                }}
                                                className="h-8"
                                                placeholder="value"
                                            />
                                        ) : (
                                            <div className="text-sm wrap-break-word pl-3">{field.value}</div>
                                        )}
                                    </div>
                                    <div className="w-[76px] flex justify-end gap-1">
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 rounded-md"
                                            title={isEditing ? "Save value" : "Edit value"}
                                            aria-label={isEditing ? "Save value" : "Edit value"}
                                            onClick={() => {
                                                if (isEditing) {
                                                    handleSaveEdit(index);
                                                    return;
                                                }

                                                handleStartEdit(index, field.value);
                                            }}
                                        >
                                            {isEditing ? <Check className="size-4" /> : <Pencil className="size-4" />}
                                        </Button>
                                        <PopConfirm
                                            title="Remove item"
                                            variant="destructive"
                                            confirmText="Remove"
                                            cancelText="Cancel"
                                            description="Are you sure you want to remove this item?"
                                            onConfirm={() => {
                                                if (editingIndex === index) {
                                                    handleCancelEdit();
                                                } else if (editingIndex !== null && editingIndex > index) {
                                                    setEditingIndex(editingIndex - 1);
                                                }

                                                remove(index);
                                            }}
                                        >
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 text-zinc-400 hover:text-red-500 hover:bg-red-50 rounded-md"
                                                title="Remove item"
                                                aria-label="Remove item"
                                            >
                                                <Trash2 className="size-4" />
                                            </Button>
                                        </PopConfirm>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </InfoBlock>
        </div>
    );
}
