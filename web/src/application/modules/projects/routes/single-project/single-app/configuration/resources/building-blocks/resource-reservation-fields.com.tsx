import { useState } from "react";

import { InputNumber } from "@components/ui/input-number";
import { useController, useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";
import { FieldListLayout } from "@application/shared/form";

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
                    stepper={0.25}
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
                    stepper={128}
                    min={0}
                    decimalScale={0}
                    fixedDecimalScale={false}
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
                <FieldListLayout
                    className="max-w-[590px]"
                    inputsClassName="grid grid-cols-2 flex-1 gap-3"
                    inputRow={
                        <>
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
                        </>
                    }
                    onAdd={() => {
                        if (!newKind.trim()) return;
                        append({ kind: newKind.trim(), value: newValue.trim() });
                        setNewKind("");
                        setNewValue("");
                    }}
                    addDisabled={!newKind.trim()}
                    items={fields.map((field, index) => ({
                        id: field.id,
                        content: (
                            <div className="grid grid-cols-2 flex-1 gap-3">
                                <span className="text-sm break-words">{field.kind}</span>
                                <span className="text-sm break-words">{field.value}</span>
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
