import { useState } from "react";

import { Button, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Plus, Trash2 } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

export function ResourceReservationFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: cpusField } = useController({ control, name: "reservations.cpus" });
    const { field: memoryField } = useController({ control, name: "reservations.memoryMB" });

    const { fields, append, remove } = useFieldArray({
        control,
        name: "reservations.genericResources",
    });

    const [newKind, setNewKind] = useState("");
    const [newValue, setNewValue] = useState("");

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="CPUs"
                        content="Number of CPUs reserved for the service."
                    />
                }
            >
                <InputNumber
                    value={cpusField.value}
                    onValueChange={val => {
                        cpusField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    min={0}
                    decimalScale={2}
                    fixedDecimalScale={false}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Memory MB"
                        content="Amount of memory (in MB) reserved for the service."
                    />
                }
            >
                <InputNumber
                    value={memoryField.value}
                    onValueChange={val => {
                        memoryField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    min={0}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Generic Resources"
                        content="User-defined resources such as GPUs or other accelerators."
                    />
                }
            >
                <div className="flex flex-col gap-3 max-w-[590px]">
                    <div className="flex gap-3 items-center">
                        <InputWithAddOn
                            addonLeft="Name"
                            value={newKind}
                            onChange={e => {
                                setNewKind(e.target.value);
                            }}
                            placeholder="SSD"
                        />
                        <InputWithAddOn
                            addonLeft="Value"
                            value={newValue}
                            onChange={e => {
                                setNewValue(e.target.value);
                            }}
                            placeholder="sda1 (string or integer)"
                        />
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => {
                                if (!newKind.trim()) return;
                                append({ kind: newKind.trim(), value: newValue.trim() });
                                setNewKind("");
                                setNewValue("");
                            }}
                            disabled={!newKind.trim()}
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>

                    <div className="divide-y divide-zinc-200">
                        {fields.map((field, index) => (
                            <div
                                key={field.id}
                                className="flex items-center gap-3 py-2"
                            >
                                <div className="grid grid-cols-2 flex-1 gap-3">
                                    <Input
                                        value={field.kind}
                                        disabled
                                    />
                                    <Input
                                        value={field.value}
                                        disabled
                                    />
                                </div>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => {
                                        remove(index);
                                    }}
                                >
                                    <Trash2 className="size-4" />
                                </Button>
                            </div>
                        ))}
                    </div>
                </div>
            </InfoBlock>
        </div>
    );
}
