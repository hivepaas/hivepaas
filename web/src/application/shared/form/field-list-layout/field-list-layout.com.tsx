import { type ReactNode } from "react";

import { Button } from "@components/ui";
import { cn } from "@lib/utils";
import { Check, Pencil, Plus, Trash2 } from "lucide-react";

import { PopConfirm } from "@application/shared/components";

export type FieldListItem = {
    id: string;
    content: ReactNode;
    onRemove: () => void;
    onEdit?: () => void;
    isEditing?: boolean;
};

type Props = {
    inputRow: ReactNode;
    onAdd: () => void;
    addDisabled?: boolean;
    disabled?: boolean;
    items: FieldListItem[];
    className?: string;
    inputsClassName?: string;
};

export function FieldListLayout({
    inputRow,
    onAdd,
    addDisabled,
    disabled = false,
    items,
    className,
    inputsClassName,
}: Props) {
    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex flex-col sm:flex-row gap-2.5 items-stretch sm:items-center">
                {inputsClassName ? <div className={inputsClassName}>{inputRow}</div> : inputRow}
                <Button
                    type="button"
                    variant="outline"
                    onClick={onAdd}
                    disabled={disabled || addDisabled}
                    className="h-9 px-4 shrink-0 w-full sm:w-auto"
                >
                    <Plus className="size-4" /> Add
                </Button>
            </div>
            <div className="mt-2 divide-y divide-zinc-200">
                {items.map(item => {
                    const hasEdit = item.onEdit !== undefined;

                    return (
                        <div
                            key={item.id}
                            className="flex items-center group gap-2 py-2"
                        >
                            {item.content}
                            <div
                                className={cn(
                                    "flex shrink-0 items-center justify-end gap-1",
                                    hasEdit ? "w-[76px]" : "w-10",
                                )}
                            >
                                {hasEdit && (
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8 text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 rounded-md"
                                        disabled={disabled}
                                        title={item.isEditing ? "Save value" : "Edit value"}
                                        aria-label={item.isEditing ? "Save value" : "Edit value"}
                                        onClick={() => {
                                            if (disabled) {
                                                return;
                                            }

                                            item.onEdit?.();
                                        }}
                                    >
                                        {item.isEditing ? <Check className="size-4" /> : <Pencil className="size-4" />}
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

                                        item.onRemove();
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
        </div>
    );
}
