import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

export function NetworksFields() {
    const { control } = useFormContext<AppConfigNetworksFormSchemaInput, unknown, AppConfigNetworksFormSchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "networkAttachments",
    });

    const [networkName, setNetworkName] = useState("");
    const [aliasesText, setAliasesText] = useState("");

    const handleAdd = () => {
        const id = networkName.trim();
        if (!id) {
            toast.error("Network name is required");
            return;
        }
        append({ id, aliasesText: aliasesText.trim() });
        setNetworkName("");
        setAliasesText("");
    };

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Networks"
                    content="Attach app to networks and optionally define network aliases."
                />
            }
        >
            <div className="flex flex-col gap-3 max-w-[560px]">
                <div className="flex gap-3 items-center">
                    <InputWithAddOn
                        addonLeft="Name"
                        value={networkName}
                        onChange={e => {
                            setNetworkName(e.target.value);
                        }}
                        placeholder="local_net_1"
                    />
                    <InputWithAddOn
                        addonLeft="Alias"
                        value={aliasesText}
                        onChange={e => {
                            setAliasesText(e.target.value);
                        }}
                        placeholder="alias1 alias2"
                    />
                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleAdd}
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
                            <Input
                                value={field.id}
                                disabled
                                className="max-w-[400px]"
                            />
                            <Input
                                value={field.aliasesText}
                                disabled
                                className="max-w-[400px]"
                            />
                            <Button
                                type="button"
                                variant="ghost"
                                className="w-[84px]"
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
    );
}
