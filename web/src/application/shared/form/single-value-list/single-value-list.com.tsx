import React, { useState } from "react";

import { Button } from "@components/ui";
import { cn } from "@lib/utils";
import { Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InputWithAddOn, PopConfirm } from "@application/shared/components";

function View<T>({ name, label = "Value", placeholder, className, checkDuplicates = true }: Props<T>) {
    const { control } = useFormContext<Record<string, { value: string }[]>>();
    const { fields, append, remove } = useFieldArray({ control, name: name as string });
    const [input, setInput] = useState("");

    return (
        <div className={cn("flex flex-col gap-2", className)}>
            <div className="flex gap-2">
                <InputWithAddOn
                    addonLeft={label}
                    value={input}
                    onChange={e => {
                        setInput(e.target.value);
                    }}
                    placeholder={placeholder ?? label}
                    className="flex-1"
                />
                <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
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
                    }}
                    disabled={input.trim() === ""}
                >
                    <Plus className="size-4" /> Add
                </Button>
            </div>
            {fields.length > 0 && (
                <div className="flex flex-col divide-y">
                    {fields.map((field, index) => (
                        <div
                            key={field.id}
                            className="flex items-center gap-2 justify-between py-1.5"
                        >
                            <div className="flex-1 grid grid-cols-1 items-center">
                                <span className="text-sm break-words">{(field as { value?: string }).value}</span>
                            </div>
                            <div className="w-[76px]">
                                <PopConfirm
                                    title="Remove item"
                                    variant="destructive"
                                    confirmText="Remove"
                                    cancelText="Cancel"
                                    description="Are you sure you want to remove this item?"
                                    onConfirm={() => {
                                        remove(index);
                                    }}
                                >
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8 text-zinc-400 hover:text-red-500 hover:bg-red-50 rounded-md"
                                    >
                                        <Trash2 className="size-3.5" />
                                    </Button>
                                </PopConfirm>
                            </div>
                        </div>
                    ))}
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
};

export const SingleValueList = React.memo(View) as typeof View;
