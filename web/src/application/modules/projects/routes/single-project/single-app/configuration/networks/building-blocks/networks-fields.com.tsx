import { useMemo, useState } from "react";

import { Button, Input } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectNetworksQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

type NetworkOptionValue = {
    id: string;
    name: string;
};

export function NetworksFields() {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { control } = useFormContext<AppConfigNetworksFormSchemaInput, unknown, AppConfigNetworksFormSchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "networkAttachments",
    });

    const [search, setSearch] = useState("");
    const [selectedNetwork, setSelectedNetwork] = useState<NetworkOptionValue | null>(null);
    const [aliasesText, setAliasesText] = useState("");

    const {
        data: { data: projectNetworks } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectNetworksQueries.useFindManyPaginated({
        projectID: projectId,
        search,
    });

    const comboboxOptions = useMemo(() => {
        return projectNetworks.map(n => ({
            value: { id: n.id, name: n.name },
            label: n.name,
        }));
    }, [projectNetworks]);

    const handleAdd = () => {
        if (!selectedNetwork) {
            toast.error("Please select a network");
            return;
        }

        console.log("selectedNetwork", selectedNetwork);
        console.log("aliasesText", aliasesText);
        append({ id: selectedNetwork.id, name: selectedNetwork.name, aliasesText: aliasesText.trim() });
        setSelectedNetwork(null);
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
                            value={selectedNetwork?.id ?? null}
                            onChange={(_, option) => {
                                setSelectedNetwork(option ?? null);
                            }}
                            onSearch={setSearch}
                            placeholder="local_net_1"
                            searchable
                            emptyText="No networks available"
                            valueKey="id"
                            loading={isFetching}
                            onRefresh={() => void refetch()}
                            isRefreshing={isRefetching}
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
