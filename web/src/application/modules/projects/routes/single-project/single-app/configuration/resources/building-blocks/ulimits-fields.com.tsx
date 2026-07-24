import { useState } from "react";

import { InputNumber } from "@components/ui/input-number";
import { useFieldArray, useFormContext } from "react-hook-form";

import { EditableCombobox, InfoBlock, InputNumberWithAddon, LabelWithInfo } from "@application/shared/components";
import { FieldListLayout } from "@application/shared/form";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

const ULIMIT_NAMES = [
    "core",
    "cpu",
    "data",
    "fsize",
    "locks",
    "memlock",
    "msgqueue",
    "nice",
    "nofile",
    "nproc",
    "rss",
    "rtprio",
    "rttime",
    "sigpending",
    "stack",
] as const;

const ulimitsGridClass = "grid grid-cols-10 flex-1 gap-2";

export function UlimitsFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { fields, append, remove, update } = useFieldArray({
        control,
        name: "ulimits",
    });

    const [name, setName] = useState<string>("");
    const [soft, setSoft] = useState<number>(0);
    const [hard, setHard] = useState<number>(0);
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftName, setDraftName] = useState("");
    const [draftSoft, setDraftSoft] = useState(0);
    const [draftHard, setDraftHard] = useState(0);

    const handleStartEdit = (index: number) => {
        const field = fields[index];
        if (!field) {
            return;
        }

        setEditingIndex(index);
        setDraftName(field.name);
        setDraftSoft(field.soft);
        setDraftHard(field.hard);
    };

    const handleSaveEdit = (index: number) => {
        if (editingIndex !== index) {
            return;
        }

        if (!draftName.trim()) {
            return;
        }

        update(index, {
            name: draftName.trim(),
            soft: draftSoft,
            hard: draftHard,
        });
        setEditingIndex(null);
        setDraftName("");
        setDraftSoft(0);
        setDraftHard(0);
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftName("");
        setDraftSoft(0);
        setDraftHard(0);
    };

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Ulimits"
                        content="Per-process resource limits for the container (e.g. nofile, nproc)."
                    />
                }
            >
                <FieldListLayout
                    className="max-w-[590px]"
                    inputsClassName={ulimitsGridClass}
                    inputRow={
                        <>
                            <div className="col-span-4 flex items-center rounded-md border border-input h-9 flex-1">
                                <span className="px-3 text-sm border-r border-input bg-muted/50 h-full flex items-center">
                                    Name
                                </span>
                                <EditableCombobox
                                    options={[...ULIMIT_NAMES]}
                                    value={name}
                                    onChange={setName}
                                    placeholder="nofile"
                                    className="-mx-px flex-1 gap-0"
                                    inputClassName="h-9 border-0 shadow-none focus-visible:ring-0 rounded-l-none"
                                />
                            </div>
                            <InputNumberWithAddon
                                addonLeft="Soft"
                                value={soft}
                                onValueChange={v => {
                                    setSoft(v ?? 0);
                                }}
                                useGrouping={false}
                                placeholder="1024"
                                classNameContainer="col-span-3"
                            />
                            <InputNumberWithAddon
                                addonLeft="Hard"
                                value={hard}
                                onValueChange={v => {
                                    setHard(v ?? 0);
                                }}
                                useGrouping={false}
                                placeholder="1024"
                                classNameContainer="col-span-3"
                            />
                        </>
                    }
                    onAdd={() => {
                        if (!name) return;
                        append({ name, soft, hard });
                        setName("");
                        setSoft(0);
                        setHard(0);
                        setEditingIndex(null);
                        setDraftName("");
                        setDraftSoft(0);
                        setDraftHard(0);
                    }}
                    addDisabled={!name}
                    items={fields.map((field, index) => {
                        const isEditing = editingIndex === index;

                        return {
                            id: field.id,
                            isEditing,
                            content: (
                                <div className={ulimitsGridClass}>
                                    {isEditing ? (
                                        <>
                                            <EditableCombobox
                                                options={[...ULIMIT_NAMES]}
                                                value={draftName}
                                                onChange={setDraftName}
                                                placeholder="nofile"
                                                className="-mx-px flex-1 gap-0 col-span-4"
                                            />
                                            <InputNumber
                                                value={draftSoft}
                                                onValueChange={v => {
                                                    setDraftSoft(v ?? 0);
                                                }}
                                                useGrouping={false}
                                                placeholder="1024"
                                                className="h-8 col-span-3"
                                            />
                                            <InputNumber
                                                value={draftHard}
                                                onValueChange={v => {
                                                    setDraftHard(v ?? 0);
                                                }}
                                                useGrouping={false}
                                                placeholder="1024"
                                                className="h-8 col-span-3"
                                            />
                                        </>
                                    ) : (
                                        <>
                                            <span className="text-sm wrap-break-word col-span-4">{field.name}</span>
                                            <span className="text-sm wrap-break-word col-span-3">{field.soft}</span>
                                            <span className="text-sm wrap-break-word col-span-3">{field.hard}</span>
                                        </>
                                    )}
                                </div>
                            ),
                            onEdit: () => {
                                if (isEditing) {
                                    handleSaveEdit(index);
                                    return;
                                }

                                handleStartEdit(index);
                            },
                            onRemove: () => {
                                if (editingIndex === index) {
                                    handleCancelEdit();
                                } else if (editingIndex !== null && editingIndex > index) {
                                    setEditingIndex(editingIndex - 1);
                                }

                                remove(index);
                            },
                        };
                    })}
                />
            </InfoBlock>
        </div>
    );
}
