import React, { useState } from "react";

import { Button, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui";
import { cn } from "@lib/utils";
import { Check, Pencil, Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InputWithAddOn, PopConfirm } from "@application/shared/components";

function View<T>({
    name,
    keyLabel = "Name",
    valueLabel = "Value",
    keyPlaceholder,
    valuePlaceholder,
    className,
    checkDuplicates = false,
    keyOptions,
    disabled = false,
    enableValueEditing = false,
}: Props<T>) {
    const { control } = useFormContext<Record<string, { key: string; value: string }[]>>();
    const { fields, append, remove, update } = useFieldArray({ control, name: name as string });

    const [keyInput, setKeyInput] = useState("");
    const [valueInput, setValueInput] = useState("");
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftValue, setDraftValue] = useState("");

    const canEditValues = enableValueEditing && !disabled;

    const handleAdd = () => {
        if (disabled) {
            return;
        }

        const trimmedKey = keyInput.trim();
        if (!trimmedKey) return;

        if (checkDuplicates) {
            const exists = fields.some(field => (field as { key?: string }).key === trimmedKey);
            if (exists) {
                toast.error(`Key "${trimmedKey}" already exists`);
                return;
            }
        }

        append({ key: trimmedKey, value: valueInput.trim() } as never);
        setKeyInput("");
        setValueInput("");
        setEditingIndex(null);
        setDraftValue("");
    };

    const handleStartEdit = (index: number, currentValue: string) => {
        if (!canEditValues) {
            return;
        }

        setEditingIndex(index);
        setDraftValue(currentValue);
    };

    const handleSaveEdit = (index: number) => {
        if (!canEditValues || editingIndex !== index) {
            return;
        }

        const currentField = fields[index] as { key?: string; value?: string };
        update(index, {
            key: currentField.key ?? "",
            value: draftValue.trim(),
        } as never);
        setEditingIndex(null);
        setDraftValue("");
    };

    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex gap-2">
                <div className="grid flex-1 grid-cols-2 gap-2">
                    {keyOptions ? (
                        <Select
                            value={keyInput}
                            onValueChange={value => {
                                if (disabled) {
                                    return;
                                }

                                setKeyInput(value);
                            }}
                            disabled={disabled}
                        >
                            <SelectTrigger>
                                <SelectValue placeholder={keyPlaceholder ?? keyLabel} />
                            </SelectTrigger>
                            <SelectContent>
                                {keyOptions.map(option => (
                                    <SelectItem
                                        key={option.value}
                                        value={option.value}
                                    >
                                        {option.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    ) : (
                        <InputWithAddOn
                            addonLeft={keyLabel}
                            value={keyInput}
                            onChange={e => {
                                if (disabled) {
                                    return;
                                }

                                setKeyInput(e.target.value);
                            }}
                            placeholder={keyPlaceholder ?? keyLabel}
                            disabled={disabled}
                        />
                    )}
                    <InputWithAddOn
                        addonLeft={valueLabel}
                        value={valueInput}
                        onChange={e => {
                            if (disabled) {
                                return;
                            }

                            setValueInput(e.target.value);
                        }}
                        placeholder={valuePlaceholder ?? valueLabel}
                        disabled={disabled}
                    />
                </div>
                <Button
                    type="button"
                    variant="outline"
                    onClick={handleAdd}
                    disabled={disabled || keyInput.trim() === ""}
                    className="h-9 px-4"
                >
                    <Plus className="size-4" /> Add
                </Button>
            </div>

            {fields.length > 0 && (
                <div className="mt-2 divide-y divide-zinc-200">
                    {fields.map((field, index) => {
                        const rowKey = (field as { key?: string }).key ?? "";
                        const rowValue = (field as { value?: string }).value ?? "";
                        const isEditing = canEditValues && editingIndex === index;

                        return (
                            <div
                                key={field.id}
                                className="flex items-center group gap-3 py-2"
                            >
                                <div className="grid grid-cols-2 flex-1 gap-3">
                                    <div className="text-sm break-words">{rowKey}</div>
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
                                                    setEditingIndex(null);
                                                    setDraftValue("");
                                                }
                                            }}
                                            placeholder={valuePlaceholder ?? valueLabel}
                                            disabled={disabled}
                                            className="h-8"
                                        />
                                    ) : (
                                        <div className="text-sm break-words">{rowValue}</div>
                                    )}
                                </div>
                                <div
                                    className={cn(
                                        "flex shrink-0 items-center justify-end gap-1",
                                        enableValueEditing ? "w-[76px]" : "w-[40px]",
                                    )}
                                >
                                    {enableValueEditing && (
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 rounded-md"
                                            disabled={disabled}
                                            title={isEditing ? "Save value" : "Edit value"}
                                            aria-label={isEditing ? "Save value" : "Edit value"}
                                            onClick={() => {
                                                if (isEditing) {
                                                    handleSaveEdit(index);
                                                    return;
                                                }

                                                handleStartEdit(index, rowValue);
                                            }}
                                        >
                                            {isEditing ? <Check className="size-4" /> : <Pencil className="size-4" />}
                                        </Button>
                                    )}
                                    <PopConfirm
                                        title="Remove item"
                                        variant="destructive"
                                        confirmText="Remove"
                                        cancelText="Cancel"
                                        description="Are you sure you want to remove this item?"
                                        onConfirm={() => {
                                            if (disabled) {
                                                return;
                                            }

                                            if (editingIndex === index) {
                                                setEditingIndex(null);
                                                setDraftValue("");
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
                                            disabled={disabled}
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
            )}
        </div>
    );
}

type Props<T> = {
    name: Path<T>;
    keyLabel?: string;
    valueLabel?: string;
    keyPlaceholder?: string;
    valuePlaceholder?: string;
    className?: string;
    checkDuplicates?: boolean;
    keyOptions?: { label: string; value: string }[];
    disabled?: boolean;
    enableValueEditing?: boolean;
};

export const KeyValueList = React.memo(View) as typeof View;
