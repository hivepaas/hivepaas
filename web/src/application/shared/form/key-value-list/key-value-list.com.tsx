import React, { useState } from "react";

import { Button, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui";
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
    enableEditing,
    enableValueEditing = false,
    keyField = "key",
    valueField = "value",
}: Props<T>) {
    const { control } = useFormContext<Record<string, Record<string, string>[]>>();
    const { fields, append, remove, update } = useFieldArray({ control, name: name as string });

    const [keyInput, setKeyInput] = useState("");
    const [valueInput, setValueInput] = useState("");
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftKey, setDraftKey] = useState("");
    const [draftValue, setDraftValue] = useState("");

    // enableValueEditing is kept as a backward-compatible alias for enableEditing.
    const showEditControls = enableEditing ?? enableValueEditing;
    const canEdit = showEditControls && !disabled;

    const getFieldValue = (field: Record<string, unknown>, fieldName: string) => {
        const value = field[fieldName];
        return typeof value === "string" ? value : "";
    };

    const handleAdd = () => {
        if (disabled) {
            return;
        }

        const trimmedKey = keyInput.trim();
        if (!trimmedKey) return;

        if (checkDuplicates) {
            const exists = fields.some(
                field => getFieldValue(field as Record<string, unknown>, keyField) === trimmedKey,
            );
            if (exists) {
                toast.error(`Key "${trimmedKey}" already exists`);
                return;
            }
        }

        append({ [keyField]: trimmedKey, [valueField]: valueInput.trim() } as never);
        setKeyInput("");
        setValueInput("");
        setEditingIndex(null);
        setDraftKey("");
        setDraftValue("");
    };

    const handleStartEdit = (index: number, currentKey: string, currentValue: string) => {
        if (!canEdit) {
            return;
        }

        setEditingIndex(index);
        setDraftKey(currentKey);
        setDraftValue(currentValue);
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftKey("");
        setDraftValue("");
    };

    const handleSaveEdit = (index: number) => {
        if (!canEdit || editingIndex !== index) {
            return;
        }

        const trimmedKey = draftKey.trim();
        if (!trimmedKey) {
            return;
        }

        if (checkDuplicates) {
            const duplicate = fields.some(
                (field, i) => i !== index && getFieldValue(field as Record<string, unknown>, keyField) === trimmedKey,
            );
            if (duplicate) {
                toast.error(`Key "${trimmedKey}" already exists`);
                return;
            }
        }

        update(index, {
            [keyField]: trimmedKey,
            [valueField]: draftValue.trim(),
        } as never);
        handleCancelEdit();
    };

    const renderKeyInput = (value: string, onChange: (next: string) => void, inputClassName?: string) => {
        if (keyOptions) {
            return (
                <Select
                    value={value}
                    onValueChange={next => {
                        if (disabled) {
                            return;
                        }

                        onChange(next);
                    }}
                    disabled={disabled}
                >
                    <SelectTrigger className={inputClassName}>
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
            );
        }

        return (
            <InputWithAddOn
                addonLeft={keyLabel}
                value={value}
                onChange={e => {
                    if (disabled) {
                        return;
                    }

                    onChange(e.target.value);
                }}
                placeholder={keyPlaceholder ?? keyLabel}
                disabled={disabled}
                // className={inputClassName}
            />
        );
    };

    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex gap-2">
                <div className="grid flex-1 grid-cols-2 gap-2">
                    {renderKeyInput(keyInput, setKeyInput)}
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
                        const row = field as Record<string, unknown>;
                        const rowKey = getFieldValue(row, keyField);
                        const rowValue = getFieldValue(row, valueField);
                        const isEditing = canEdit && editingIndex === index;

                        return (
                            <div
                                key={field.id}
                                className="flex items-center group gap-2 py-2"
                            >
                                <div className="grid grid-cols-2 flex-1 gap-2">
                                    {isEditing ? (
                                        <>
                                            {renderKeyInput(draftKey, setDraftKey, "h-8")}
                                            <InputWithAddOn
                                                addonLeft={valueLabel}
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
                                                placeholder={valuePlaceholder ?? valueLabel}
                                                disabled={disabled}
                                            />
                                        </>
                                    ) : (
                                        <>
                                            <div className="text-sm wrap-break-word">{rowKey}</div>
                                            <div className="text-sm wrap-break-word">{rowValue}</div>
                                        </>
                                    )}
                                </div>
                                <div
                                    className={cn(
                                        "flex shrink-0 items-center justify-end gap-1",
                                        showEditControls ? "w-[76px]" : "w-10",
                                    )}
                                >
                                    {showEditControls && (
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 rounded-md"
                                            disabled={disabled}
                                            title={isEditing ? "Save" : "Edit"}
                                            aria-label={isEditing ? "Save" : "Edit"}
                                            onClick={() => {
                                                if (isEditing) {
                                                    handleSaveEdit(index);
                                                    return;
                                                }

                                                handleStartEdit(index, rowKey, rowValue);
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
    /** Enables inline editing of both key and value columns. */
    enableEditing?: boolean;
    /** Backward-compatible alias for `enableEditing`. */
    enableValueEditing?: boolean;
    keyField?: string;
    valueField?: string;
};

export const KeyValueList = React.memo(View) as typeof View;
