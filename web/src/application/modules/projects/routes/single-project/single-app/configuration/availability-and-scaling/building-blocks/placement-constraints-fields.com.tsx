import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Check, Pencil, Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";
import { EAppServicePlacement } from "~/projects/module-shared/enums";

import {
    EditComboboxWithAddOn,
    EditableCombobox,
    InfoBlock,
    LabelWithInfo,
    PopConfirm,
} from "@application/shared/components";

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

    const { fields, append, remove, update } = useFieldArray({
        control,
        name: "constraints",
    });

    const [newName, setNewName] = useState<string>(CONSTRAINT_NAME_OPTIONS[0]);
    const [newOp, setNewOp] = useState<EAppServicePlacement>(EAppServicePlacement.Equal);
    const [newValue, setNewValue] = useState("");
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftName, setDraftName] = useState("");
    const [draftOp, setDraftOp] = useState<EAppServicePlacement>(EAppServicePlacement.Equal);
    const [draftValue, setDraftValue] = useState("");

    const handleAddConstraint = () => {
        if (newName && newValue) {
            append({ name: newName, op: newOp, value: newValue });
            setNewName(CONSTRAINT_NAME_OPTIONS[0]);
            setNewOp(EAppServicePlacement.Equal);
            setNewValue("");
            setEditingIndex(null);
        } else {
            toast.error("Please fill in all constraint fields");
        }
    };

    const handleStartEdit = (index: number) => {
        const field = fields[index];
        if (!field) {
            return;
        }

        setEditingIndex(index);
        setDraftName(field.name);
        setDraftOp(field.op);
        setDraftValue(field.value);
    };

    const handleSaveEdit = (index: number) => {
        if (editingIndex !== index) {
            return;
        }

        if (!draftName.trim() || !draftValue.trim()) {
            toast.error("Please fill in all constraint fields");
            return;
        }

        update(index, {
            name: draftName.trim(),
            op: draftOp,
            value: draftValue.trim(),
        });
        setEditingIndex(null);
        setDraftName("");
        setDraftValue("");
        setDraftOp(EAppServicePlacement.Equal);
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftName("");
        setDraftValue("");
        setDraftOp(EAppServicePlacement.Equal);
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
                        <div
                            className="grid gap-3 flex-1"
                            style={{ gridTemplateColumns: "1fr 120px 1fr" }}
                        >
                            <EditComboboxWithAddOn
                                addonLeft="Name"
                                options={[...CONSTRAINT_NAME_OPTIONS]}
                                value={newName}
                                onChange={setNewName}
                                placeholder="Select or type name"
                                inputClassName="h-9"
                            />

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
                                    <SelectTrigger className="flex-1 border-0 shadow-none focus:ring-0 rounded-l-none h-9">
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
                        </div>

                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAddConstraint}
                            disabled={!newName || !newValue}
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
                                    <div className="grid grid-cols-[1fr_120px_1fr] flex-1 gap-3">
                                        {isEditing ? (
                                            <>
                                                <EditableCombobox
                                                    options={[...CONSTRAINT_NAME_OPTIONS]}
                                                    value={draftName}
                                                    onChange={setDraftName}
                                                    placeholder="Select or type name"
                                                    inputClassName="h-8"
                                                />
                                                <div className="flex items-center rounded-md shadow-xs bg-background border border-input">
                                                    <span className="px-3 text-sm border-r border-input bg-muted/50 rounded-l-md h-8 flex items-center">
                                                        Op
                                                    </span>
                                                    <Select
                                                        value={draftOp}
                                                        onValueChange={v => {
                                                            setDraftOp(v as EAppServicePlacement);
                                                        }}
                                                    >
                                                        <SelectTrigger className="flex-1 border-0 shadow-none focus:ring-0 rounded-l-none h-8">
                                                            <SelectValue placeholder="Select op" />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            <SelectItem value={EAppServicePlacement.Equal}>
                                                                ==
                                                            </SelectItem>
                                                            <SelectItem value={EAppServicePlacement.NotEqual}>
                                                                !=
                                                            </SelectItem>
                                                        </SelectContent>
                                                    </Select>
                                                </div>
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
                                                    placeholder="value"
                                                />
                                            </>
                                        ) : (
                                            <>
                                                <div className="text-sm wrap-break-word">{field.name}</div>
                                                <div className="text-sm font-mono">{field.op}</div>
                                                <div className="text-sm wrap-break-word">{field.value}</div>
                                            </>
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

                                                handleStartEdit(index);
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
