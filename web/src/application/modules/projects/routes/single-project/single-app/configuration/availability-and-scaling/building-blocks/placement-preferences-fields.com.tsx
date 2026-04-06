import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigAvailabilitySchemaInput, type AppConfigAvailabilitySchemaOutput } from "../schemas";

export function PlacementPreferencesFields() {
    const { control } = useFormContext<AppConfigAvailabilitySchemaInput, unknown, AppConfigAvailabilitySchemaOutput>();

    const { fields, append, remove } = useFieldArray({
        control,
        name: "preferences",
    });

    const [newValue, setNewValue] = useState("");
    const strategy = "spread";

    const handleAddPreference = () => {
        if (newValue) {
            append({ name: strategy, value: newValue });
            setNewValue("");
        } else {
            toast.error("Please provide a value for the placement preference");
        }
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
                        {fields.map((field, index) => (
                            <div
                                key={field.id}
                                className="flex items-center group gap-4 py-2"
                            >
                                <div className="grid grid-cols-2 flex-1 gap-3">
                                    <div className="text-sm break-words pl-3">{field.name}</div>
                                    <div className="text-sm break-words pl-3">{field.value}</div>
                                </div>
                                <div className="w-[76px] flex justify-start">
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        onClick={() => {
                                            remove(index);
                                        }}
                                        className="h-8 w-8 text-zinc-400 hover:text-red-500 hover:bg-red-50 rounded-md"
                                    >
                                        <Trash2 className="size-4" />
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </InfoBlock>
        </div>
    );
}
