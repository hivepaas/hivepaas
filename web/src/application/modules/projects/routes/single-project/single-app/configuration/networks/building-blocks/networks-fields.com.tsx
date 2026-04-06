import { useMemo, useState } from "react";

import { Button, Input } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { Combobox, InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

/** Mock networks; replace with API-driven list when available. */
const MOCK_PROJECT_NETWORKS = [
    { id: "net_local_1", name: "local_net_1" },
    { id: "net_local_2", name: "local_net_2" },
    { id: "net_bridge", name: "bridge" },
    { id: "net_overlay", name: "overlay_net" },
] as const;

type NetworkOptionValue = {
    id: string;
    name: string;
};

export function NetworksFields() {
    const { control } = useFormContext<AppConfigNetworksFormSchemaInput, unknown, AppConfigNetworksFormSchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "networkAttachments",
    });

    const [search, setSearch] = useState("");
    const [selectedNetworkId, setSelectedNetworkId] = useState<string | null>(null);
    const [aliasesText, setAliasesText] = useState("");

    const comboboxOptions = useMemo(() => {
        const q = search.trim().toLowerCase();
        return MOCK_PROJECT_NETWORKS.filter(
            n => !q || n.id.toLowerCase().includes(q) || n.name.toLowerCase().includes(q),
        ).map(n => ({
            value: { id: n.id, name: n.name },
            label: n.name,
        }));
    }, [search]);

    const handleAdd = () => {
        if (!selectedNetworkId) {
            toast.error("Please select a network");
            return;
        }
        const net = MOCK_PROJECT_NETWORKS.find(n => n.id === selectedNetworkId);
        if (!net) {
            toast.error("Invalid network selection");
            return;
        }
        append({ id: net.id, name: net.name, aliasesText: aliasesText.trim() });
        setSelectedNetworkId(null);
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
            <div className="flex flex-col gap-3 ">
                <div className="flex gap-3 items-center">
                    <div className="grid flex-1 grid-cols-2 gap-3 max-w-[500px]">
                        <Combobox<NetworkOptionValue>
                            options={comboboxOptions}
                            value={selectedNetworkId}
                            onChange={id => {
                                setSelectedNetworkId(id);
                            }}
                            onSearch={setSearch}
                            placeholder="local_net_1"
                            searchable
                            emptyText="No networks match your search"
                            valueKey="id"
                        />
                        <InputWithAddOn
                            addonLeft="Alias"
                            value={aliasesText}
                            onChange={e => {
                                setAliasesText(e.target.value);
                            }}
                            placeholder="alias1 alias2"
                        />
                    </div>
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
                            <div className="grid grid-cols-2 flex-1 gap-3 max-w-[500px]">
                                <Input
                                    value={field.name || field.id}
                                    disabled
                                    className="max-w-[400px]"
                                />
                                <Input
                                    value={field.aliasesText}
                                    disabled
                                    className="max-w-[400px]"
                                />
                            </div>
                            <Button
                                type="button"
                                variant="ghost"
                                className="w-[76px]"
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
