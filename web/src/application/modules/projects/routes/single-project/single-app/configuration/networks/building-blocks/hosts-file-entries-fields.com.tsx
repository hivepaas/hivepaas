import { useState } from "react";

import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";
import { FieldListLayout } from "@application/shared/form";

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
            <FieldListLayout
                inputsClassName="grid flex-1 grid-cols-2 gap-3 max-w-[500px]"
                inputRow={
                    <>
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
                    </>
                }
                onAdd={handleAdd}
                addDisabled={address.trim() === ""}
                items={fields.map((field, index) => ({
                    id: field.id,
                    content: (
                        <div className="grid grid-cols-2 flex-1 gap-3 max-w-[500px]">
                            <span className="text-sm break-words">{field.address}</span>
                            <span className="text-sm break-words">{field.hostnamesText}</span>
                        </div>
                    ),
                    onRemove: () => {
                        remove(index);
                    },
                }))}
            />
        </InfoBlock>
    );
}
