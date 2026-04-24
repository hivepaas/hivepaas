import { useState } from "react";

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

export function UlimitsFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { fields, append, remove } = useFieldArray({
        control,
        name: "ulimits",
    });

    const [name, setName] = useState<string>("");
    const [soft, setSoft] = useState<number>(0);
    const [hard, setHard] = useState<number>(0);

    return (
        <div className="flex flex-col gap-6 px-2">
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
                    inputsClassName="grid grid-cols-10 flex-1 gap-3"
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
                    }}
                    addDisabled={!name}
                    items={fields.map((field, index) => ({
                        id: field.id,
                        content: (
                            <div className="grid grid-cols-10 flex-1 gap-3">
                                <span className="text-sm break-words col-span-4">{field.name}</span>
                                <span className="text-sm break-words col-span-3">{field.soft}</span>
                                <span className="text-sm break-words col-span-3">{field.hard}</span>
                            </div>
                        ),
                        onRemove: () => {
                            remove(index);
                        },
                    }))}
                />
            </InfoBlock>
        </div>
    );
}
