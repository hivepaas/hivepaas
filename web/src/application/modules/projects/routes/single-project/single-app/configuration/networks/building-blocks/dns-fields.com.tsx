import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

type DnsListName = "dnsConfig.nameservers" | "dnsConfig.search" | "dnsConfig.options";

function DNSListFields({
    label,
    listName,
    placeholder,
    content,
}: {
    label: string;
    listName: DnsListName;
    placeholder: string;
    content: string;
}) {
    const { control } = useFormContext<AppConfigNetworksFormSchemaInput, unknown, AppConfigNetworksFormSchemaOutput>();
    const [value, setInputValue] = useState("");

    const { fields, append, remove } = useFieldArray({
        control,
        name: listName,
    });

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label={label}
                    content={content}
                />
            }
        >
            <div className="flex flex-col gap-3 max-w-[590px]">
                <div className="flex gap-3 items-center">
                    <InputWithAddOn
                        addonLeft="Value"
                        value={value}
                        onChange={e => {
                            setInputValue(e.target.value);
                        }}
                        placeholder={placeholder}
                    />
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            const next = value.trim();
                            if (!next) return;
                            append({ value: next });
                            setInputValue("");
                        }}
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
                                value={field.value}
                                disabled
                            />
                            <div className="w-[76px]">
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
                        </div>
                    ))}
                </div>
            </div>
        </InfoBlock>
    );
}

export function DNSFields() {
    return (
        <div className="flex flex-col gap-6">
            <DNSListFields
                label="DNS: Nameservers"
                listName="dnsConfig.nameservers"
                placeholder="nameserver"
                content="Add values for DNS nameservers."
            />
            <DNSListFields
                label="DNS: Search"
                listName="dnsConfig.search"
                placeholder="search"
                content="Add values for DNS search."
            />
            <DNSListFields
                label="DNS: Options"
                listName="dnsConfig.options"
                placeholder="option"
                content="Add values for DNS options."
            />
        </div>
    );
}
