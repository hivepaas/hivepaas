import React, { useState } from "react";

import { Button } from "@components/ui";
import { cn } from "@lib/utils";
import { Plus, Trash2 } from "lucide-react";
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
}: Props<T>) {
    const { control } = useFormContext<Record<string, { key: string; value: string }[]>>();
    const { fields, append, remove } = useFieldArray({ control, name: name as string });

    const [keyInput, setKeyInput] = useState("");
    const [valueInput, setValueInput] = useState("");

    const handleAdd = () => {
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
    };

    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex gap-2">
                <div className="grid flex-1 grid-cols-2 gap-2">
                    <InputWithAddOn
                        addonLeft={keyLabel}
                        value={keyInput}
                        onChange={e => {
                            setKeyInput(e.target.value);
                        }}
                        placeholder={keyPlaceholder ?? keyLabel}
                    />
                    <InputWithAddOn
                        addonLeft={valueLabel}
                        value={valueInput}
                        onChange={e => {
                            setValueInput(e.target.value);
                        }}
                        placeholder={valuePlaceholder ?? valueLabel}
                    />
                </div>
                <Button
                    type="button"
                    variant="outline"
                    onClick={handleAdd}
                    disabled={keyInput.trim() === ""}
                    className="h-9 px-4"
                >
                    <Plus className="size-4" /> Add
                </Button>
            </div>

            {fields.length > 0 && (
                <div className="mt-2 divide-y divide-zinc-200">
                    {fields.map((field, index) => (
                        <div
                            key={field.id}
                            className="flex items-center group gap-3 py-2"
                        >
                            <div className="grid grid-cols-2 flex-1 gap-3">
                                <div className="text-sm break-words">{(field as { key?: string }).key}</div>
                                <div className="text-sm break-words">{(field as { value?: string }).value}</div>
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
                                        <Trash2 className="size-4" />
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
    keyLabel?: string;
    valueLabel?: string;
    keyPlaceholder?: string;
    valuePlaceholder?: string;
    className?: string;
    checkDuplicates?: boolean;
};

export const KeyValueList = React.memo(View) as typeof View;
