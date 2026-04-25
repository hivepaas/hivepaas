import { useState } from "react";

import { Input } from "@components/ui";
import { Checkbox } from "@components/ui/checkbox";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputNumberWithAddon, InputWithAddOn, LabelWithInfo } from "@application/shared/components";
import { FieldListLayout } from "@application/shared/form";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

export function CapabilitiesFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: capabilityAddField } = useController({ control, name: "capabilities.capabilityAdd" });
    const { field: capabilityDropField } = useController({ control, name: "capabilities.capabilityDrop" });
    const { field: enableGPUField } = useController({ control, name: "capabilities.enableGPU" });
    const { field: oomScoreAdjField } = useController({ control, name: "capabilities.oomScoreAdj" });

    const { fields, append, remove } = useFieldArray({
        control,
        name: "capabilities.sysctls",
    });

    const [newName, setNewName] = useState("");
    const [newValue, setNewValue] = useState("");

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Capabilities Add"
                        content="Linux capabilities to add to the container."
                    />
                }
            >
                <Input
                    value={capabilityAddField.value}
                    onChange={e => {
                        capabilityAddField.onChange(e.target.value);
                    }}
                    placeholder="SYS_ADMIN ANOTHER"
                    className="max-w-[500px]"
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Capabilities Drop"
                        content="Linux capabilities to drop from the container."
                    />
                }
            >
                <Input
                    value={capabilityDropField.value}
                    onChange={e => {
                        capabilityDropField.onChange(e.target.value);
                    }}
                    placeholder="AUDIT_WRITE ANOTHER"
                    className="max-w-[500px]"
                />
            </InfoBlock>

            <InfoBlock title="Enable GPU">
                <div className="flex items-center gap-2">
                    <Checkbox
                        checked={enableGPUField.value}
                        onCheckedChange={val => {
                            enableGPUField.onChange(val);
                        }}
                    />
                    <span className="text-sm text-muted-foreground">
                        This will add <code className="text-orange-500">[gpu]</code> to capabilities, see{" "}
                        <a
                            href="https://docs.docker.com/compose/how-tos/gpu-support/"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-500 underline underline-offset-2"
                        >
                            docs
                        </a>
                    </span>
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="OomScoreAdj"
                        content="Adjusts the OOM killer score for the container process."
                    />
                }
            >
                <InputNumber
                    value={oomScoreAdjField.value}
                    onValueChange={val => {
                        oomScoreAdjField.onChange(val);
                    }}
                    className="max-w-[100px]"
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Sysctls"
                        content="Kernel parameters to set in the container namespace."
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
                                value={newName}
                                onChange={e => {
                                    setNewName(e.target.value);
                                }}
                                placeholder="net.core.somaxconn"
                            />
                            <InputNumberWithAddon
                                addonLeft="Value"
                                value={
                                    newValue.trim() === "" || Number.isNaN(Number(newValue))
                                        ? undefined
                                        : Number(newValue)
                                }
                                onValueChange={v => {
                                    setNewValue(v === undefined ? "" : String(v));
                                }}
                                useGrouping={false}
                                placeholder="1024"
                            />
                        </>
                    }
                    onAdd={() => {
                        if (!newName.trim()) return;
                        append({ name: newName.trim(), value: newValue.trim() });
                        setNewName("");
                        setNewValue("");
                    }}
                    addDisabled={!newName.trim()}
                    items={fields.map((field, index) => ({
                        id: field.id,
                        content: (
                            <div className="grid grid-cols-2 flex-1 gap-3">
                                <span className="text-sm break-words">{field.name}</span>
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
