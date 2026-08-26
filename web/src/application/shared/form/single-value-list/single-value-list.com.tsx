import React, { useState } from "react";

import { Button } from "@components/ui";
import { cn } from "@lib/utils";
import { Check, Pencil, Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InputWithAddOn, PopConfirm } from "@application/shared/components";

function View<T>({
    name,
    label = "Value",
    placeholder,
    className,
    checkDuplicates = true,
    disabled = false,
    enableEditing,
    enableValueEditing = false,
}: Props<T>) {
    const { control } = useFormContext<Record<string, { value: string }[]>>();
    const { fields, append, remove, update } = useFieldArray({ control, name: name as string });
    const [input, setInput] = useState("");
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftValue, setDraftValue] = useState("");

    // enableValueEditing is kept as a backward-compatible alias for enableEditing.
    const showEditControls = enableEditing ?? enableValueEditing;
    const canEdit = showEditControls && !disabled;

    const handleAdd = () => {
        if (disabled) {
            return;
        }

        const trimmed = input.trim();
        if (!trimmed) return;
        if (checkDuplicates) {
            const exists = fields.some(f => (f as { value?: string }).value === trimmed);
            if (exists) {
                toast.error(`"${trimmed}" already exists`);
                return;
            }
        }
        append({ value: trimmed } as never);
        setInput("");
        setEditingIndex(null);
        setDraftValue("");
    };

    const handleStartEdit = (index: number, currentValue: string) => {
        if (!canEdit) {
            return;
        }

        setEditingIndex(index);
        setDraftValue(currentValue);
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftValue("");
    };

    const handleSaveEdit = (index: number) => {
        if (!canEdit || editingIndex !== index) {
            return;
        }

        const trimmed = draftValue.trim();
        if (!trimmed) {
            return;
        }

        if (checkDuplicates) {
            const duplicate = fields.some((field, i) => i !== index && (field as { value?: string }).value === trimmed);
            if (duplicate) {
                toast.error(`"${trimmed}" already exists`);
                return;
            }
        }

        update(index, { value: trimmed } as never);
        handleCancelEdit();
    };

    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex gap-2">
                <InputWithAddOn
                    addonLeft={label}
                    value={input}
                    onChange={e => {
                        if (disabled) {
                            return;
                        }

                        setInput(e.target.value);
                    }}
                    onKeyDown={e => {
                        if (e.key === "Enter") {
                            e.preventDefault();
                            handleAdd();
                        }
                    }}
                    placeholder={placeholder ?? label}
                    className="flex-1"
                    disabled={disabled}
                />
                <Button
                    type="button"
                    variant="outline"
                    onClick={handleAdd}
                    disabled={disabled || input.trim() === ""}
                    className="h-9 px-4"
                >
                    <Plus className="size-4" /> Add
                </Button>
            </div>
            {fields.length > 0 && (
                <div className="mt-2 divide-y divide-zinc-200">
                    {fields.map((field, index) => {
                        const value = (field as { value?: string }).value ?? "";
                        const isEditing = canEdit && editingIndex === index;

                        return (
                            <div
                                key={field.id}
                                className="flex items-center group gap-2 py-2"
                            >
                                <div className="flex-1 min-w-0">
                                    {isEditing ? (
                                        <InputWithAddOn
                                            addonLeft={label}
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
                                            placeholder={placeholder ?? label}
                                            disabled={disabled}
                                        />
                                    ) : (
                                        <div className="text-sm wrap-break-word break-words min-w-0">{value}</div>
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

                                                handleStartEdit(index, value);
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
    label?: string;
    placeholder?: string;
    className?: string;
    checkDuplicates?: boolean;
    disabled?: boolean;
    /** Enables inline editing of values. */
    enableEditing?: boolean;
    /** Backward-compatible alias for `enableEditing`. */
    enableValueEditing?: boolean;
};

export const SingleValueList = React.memo(View) as typeof View;
