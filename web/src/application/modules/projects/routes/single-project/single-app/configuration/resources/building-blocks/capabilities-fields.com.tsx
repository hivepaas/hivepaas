import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Checkbox } from "@components/ui/checkbox";
import { InputNumber } from "@components/ui/input-number";
import { Plus, Trash2 } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

export function CapabilitiesFields() {
    const { control, register } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: capabilityAddField } = useController({ control, name: "capabilities.capabilityAdd" });
    const { field: capabilityDropField } = useController({ control, name: "capabilities.capabilityDrop" });
    const { field: enableGPUField } = useController({ control, name: "capabilities.enableGPU" });
    const { field: oomScoreAdjField } = useController({ control, name: "capabilities.oomScoreAdj" });
    const { field: sysctlsField } = useController({ control, name: "capabilities.sysctls" });

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
                    {...register("capabilities.capabilityAdd")}
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
                    {...register("capabilities.capabilityDrop")}
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
                            href="https://docs.docker.com/reference/api/engine/version/v1.52/#tag/Service/operation/ServiceUpdate"
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
                <div className="flex flex-col gap-3 max-w-[590px]">
                    <div className="flex gap-3 items-center">
                        <InputWithAddOn
                            addonLeft="Name"
                            value={newName}
                            onChange={e => {
                                setNewName(e.target.value);
                            }}
                            placeholder="net.core.somaxconn"
                        />
                        <InputWithAddOn
                            addonLeft="Value"
                            value={newValue}
                            onChange={e => {
                                setNewValue(e.target.value);
                            }}
                            placeholder="1024"
                        />
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => {
                                if (!newName.trim()) return;
                                append({ name: newName.trim(), value: newValue.trim() });
                                setNewName("");
                                setNewValue("");
                            }}
                            disabled={!newName.trim()}
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
                                        value={field.name}
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
