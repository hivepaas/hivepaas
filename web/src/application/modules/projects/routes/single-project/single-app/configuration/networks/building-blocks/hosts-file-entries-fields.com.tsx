import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

export function HostsFileEntriesFields() {
    const { control } = useFormContext<AppConfigNetworksFormSchemaInput, unknown, AppConfigNetworksFormSchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "hostsFileEntries",
    });

    const [address, setAddress] = useState("");
    const [hostnamesText, setHostnamesText] = useState("");

    const handleAdd = () => {
        if (!address.trim()) {
            toast.error("Host address is required");
            return;
        }
        append({
            address: address.trim(),
            hostnamesText: hostnamesText.trim(),
        });
        setAddress("");
        setHostnamesText("");
    };

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Hosts file entries"
                    content="Static host mappings written to container hosts file."
                />
            }
        >
            <div className="flex flex-col gap-3 ">
                <div className="flex gap-3 items-center">
                    <div className="grid flex-1 grid-cols-2 gap-3 max-w-[500px]">
                        <InputWithAddOn
                            addonLeft="Addr"
                            value={address}
                            onChange={e => {
                                setAddress(e.target.value);
                            }}
                            placeholder="11.22.33.44"
                        />
                        <InputWithAddOn
                            addonLeft="Name"
                            value={hostnamesText}
                            onChange={e => {
                                setHostnamesText(e.target.value);
                            }}
                            placeholder="hostname alias1 alias2"
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
                                    value={field.address}
                                    disabled
                                    className="max-w-[400px]"
                                />
                                <Input
                                    value={field.hostnamesText}
                                    disabled
                                    className="max-w-[400px]"
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
    );
}
