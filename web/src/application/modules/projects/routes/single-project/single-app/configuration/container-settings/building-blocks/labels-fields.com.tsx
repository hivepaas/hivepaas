import { useState } from "react";

import { Button } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

export function LabelsFields() {
    const { control } = useFormContext<
        AppConfigContainerSettingsFormSchemaInput,
        unknown,
        AppConfigContainerSettingsFormSchemaOutput
    >();

    const { fields, append, remove } = useFieldArray({
        control,
        name: "labels",
    });

    const [newLabelKey, setNewLabelKey] = useState("");
    const [newLabelValue, setNewLabelValue] = useState("");

    const handleAddLabel = () => {
        if (newLabelKey && newLabelValue) {
            const exists = fields.some(field => field.key === newLabelKey);

            if (exists) {
                toast.error(`Label key "${newLabelKey}" already exists`);
                return;
            }

            append({ key: newLabelKey, value: newLabelValue });
            setNewLabelKey("");
            setNewLabelValue("");
        }
    };

    return (
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Labels</h3>
            <div className="flex flex-col gap-6 px-2">
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Labels"
                            content="Key-value labels applied to the container."
                        />
                    }
                >
                    <div className="flex flex-col gap-4 max-w-[660px]">
                        <div className="flex gap-4 flex-wrap">
                            <InputWithAddOn
                                addonLeft="Name"
                                value={newLabelKey}
                                onChange={e => {
                                    setNewLabelKey(e.target.value);
                                }}
                            />
                            <InputWithAddOn
                                addonLeft="Value"
                                value={newLabelValue}
                                onChange={e => {
                                    setNewLabelValue(e.target.value);
                                }}
                            />
                            <Button
                                type="button"
                                variant="outline"
                                onClick={handleAddLabel}
                                disabled={newLabelKey === "" || newLabelValue === ""}
                                className="h-9 px-4"
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
                                    <div className="grid grid-cols-2 flex-1 gap-4">
                                        <div className="text-sm break-words">{field.key}</div>
                                        <div className="text-sm break-words">{field.value}</div>
                                    </div>
                                    <div className="w-[84px]">
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
        </>
    );
}
