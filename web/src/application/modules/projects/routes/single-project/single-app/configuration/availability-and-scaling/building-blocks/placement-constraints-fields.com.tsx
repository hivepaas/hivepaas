import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";
import { EAppServicePlacement } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigAvailabilitySchemaInput, type AppConfigAvailabilitySchemaOutput } from "../schemas";

const CONSTRAINT_NAME_OPTIONS = [
    "node.id",
    "node.name",
    "node.hostname",
    "node.role",
    "node.platform.os",
    "node.platform.arch",
    "node.labels.xxx",
    "engine.labels.xxx",
] as const;

export function PlacementConstraintsFields() {
    const { control } = useFormContext<AppConfigAvailabilitySchemaInput, unknown, AppConfigAvailabilitySchemaOutput>();

    const { fields, append, remove } = useFieldArray({
        control,
        name: "constraints",
    });

    const [newName, setNewName] = useState<string>(CONSTRAINT_NAME_OPTIONS[0]);
    const [newOp, setNewOp] = useState<EAppServicePlacement>(EAppServicePlacement.Equal);
    const [newValue, setNewValue] = useState("");

    const handleAddConstraint = () => {
        if (newName && newValue) {
            append({ name: newName, op: newOp, value: newValue });
            setNewName(CONSTRAINT_NAME_OPTIONS[0]);
            setNewOp(EAppServicePlacement.Equal);
            setNewValue("");
        } else {
            toast.error("Please fill in all constraint fields");
        }
    };

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Placement Constraints"
                        content="Constraints that must be met for the service to be scheduled on a node."
                    />
                }
            >
                <div className="flex flex-col gap-4 max-w-[800px]">
                    <div className="flex gap-4 flex-wrap items-center">
                        <div className="flex items-center rounded-md shadow-xs bg-background border border-input">
                            <span className="px-3 text-sm border-r border-input bg-muted/50 rounded-l-md h-9 flex items-center">
                                Name
                            </span>
                            <Select
                                value={newName}
                                onValueChange={setNewName}
                            >
                                <SelectTrigger className="w-[180px] border-0 shadow-none focus:ring-0 rounded-l-none h-9">
                                    <SelectValue placeholder="Select name" />
                                </SelectTrigger>
                                <SelectContent>
                                    {CONSTRAINT_NAME_OPTIONS.map(opt => (
                                        <SelectItem
                                            key={opt}
                                            value={opt}
                                        >
                                            {opt}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="flex items-center rounded-md shadow-xs bg-background border border-input">
                            <span className="px-3 text-sm border-r border-input bg-muted/50 rounded-l-md h-9 flex items-center">
                                Op
                            </span>
                            <Select
                                value={newOp}
                                onValueChange={v => {
                                    setNewOp(v as EAppServicePlacement);
                                }}
                            >
                                <SelectTrigger className="w-[80px] border-0 shadow-none focus:ring-0 rounded-l-none h-9">
                                    <SelectValue placeholder="Select op" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={EAppServicePlacement.Equal}>==</SelectItem>
                                    <SelectItem value={EAppServicePlacement.NotEqual}>!=</SelectItem>
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

                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAddConstraint}
                            disabled={!newName || !newValue}
                            className="h-9 px-4 shrink-0"
                        >
                            <Plus className="size-4 mr-2" /> Add
                        </Button>
                    </div>

                    <div className="mt-2 divide-y divide-zinc-200">
                        {fields.map((field, index) => (
                            <div
                                key={field.id}
                                className="flex items-center group gap-4 py-2"
                            >
                                <div className="grid grid-cols-[1fr_80px_1fr] flex-1 gap-4 items-center">
                                    <div className="text-sm break-words">{field.name}</div>
                                    <div className="text-sm font-mono text-center">{field.op}</div>
                                    <div className="text-sm break-words">{field.value}</div>
                                </div>
                                <div className="w-[84px] flex justify-end">
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
