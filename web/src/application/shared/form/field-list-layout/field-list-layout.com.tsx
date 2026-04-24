import { type ReactNode } from "react";

import { Button } from "@components/ui";
import { cn } from "@lib/utils";
import { Plus, Trash2 } from "lucide-react";

import { PopConfirm } from "@application/shared/components";

export type FieldListItem = {
    id: string;
    content: ReactNode;
    onRemove: () => void;
};

type Props = {
    inputRow: ReactNode;
    onAdd: () => void;
    addDisabled?: boolean;
    items: FieldListItem[];
    className?: string;
    inputsClassName?: string;
};

export function FieldListLayout({ inputRow, onAdd, addDisabled, items, className, inputsClassName }: Props) {
    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex gap-3 items-center">
                {inputsClassName ? <div className={inputsClassName}>{inputRow}</div> : inputRow}
                <Button
                    type="button"
                    variant="outline"
                    onClick={onAdd}
                    disabled={addDisabled}
                >
                    <Plus className="size-4" /> Add
                </Button>
            </div>
            <div className="divide-y divide-zinc-200">
                {items.map(item => (
                    <div
                        key={item.id}
                        className="flex items-center gap-3 py-2"
                    >
                        {item.content}
                        <div className="w-[76px]">
                            <PopConfirm
                                title="Remove item"
                                variant="destructive"
                                confirmText="Remove"
                                cancelText="Cancel"
                                description="Are you sure you want to remove this item?"
                                onConfirm={item.onRemove}
                            >
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 text-zinc-400 hover:text-red-500 hover:bg-red-50 rounded-md"
                                >
                                    <Trash2 className="size-4" />
                                </Button>
                            </PopConfirm>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
