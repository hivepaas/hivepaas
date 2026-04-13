import { useMemo, useState } from "react";

import { Button } from "@components/ui";
import { SelectItem } from "@components/ui/select";
import { Plus, RefreshCw, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import { ComboboxWithAddon, InfoBlock, InputWithAddOn, LabelWithInfo, SelectWithAddon } from "@application/shared/components";

import { STORAGE_CLUSTER_SOURCE_OPTIONS_MOCK, STORAGE_VOLUME_SOURCE_OPTIONS_MOCK } from "../constants";
import { type AppConfigStorageFormSchemaInput, type AppConfigStorageFormSchemaOutput } from "../schemas";

type SourceOptionValue = { name: string };

function isComboboxSourceType(type: EMountType) {
    return type === EMountType.Volume || type === EMountType.Cluster;
}

export function StorageMountsFields() {
    const { control } = useFormContext<AppConfigStorageFormSchemaInput, unknown, AppConfigStorageFormSchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "mounts",
    });

    const [type, setType] = useState<EMountType>(EMountType.Bind);
    const [source, setSource] = useState("");
    const [target, setTarget] = useState("");
    const [consistency, setConsistency] = useState<EMountConsistency>(EMountConsistency.Default);

    const sourceOptions = useMemo(() => {
        const names =
            type === EMountType.Cluster ? STORAGE_CLUSTER_SOURCE_OPTIONS_MOCK : STORAGE_VOLUME_SOURCE_OPTIONS_MOCK;
        return names.map(name => ({
            label: name,
            value: { name },
        }));
    }, [type]);

    const isTmpfs = type === EMountType.Tmpfs;
    const requiresSource = !isTmpfs;
    const canAdd = target.trim().length > 0 && (!requiresSource || source.trim().length > 0);

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Mounts"
                    content="Configure mount type, source, target path, and consistency for this app."
                />
            }
            titleWidth={150}
        >
            <div className="flex flex-col gap-3">
                <div className="flex items-center gap-2">
                    <div className="grid flex-1 grid-cols-8 gap-3">
                        <SelectWithAddon
                            addonLeft="Type"
                            value={type}
                            onValueChange={value => {
                                const next = value as EMountType;
                                setType(next);
                                if (next === EMountType.Tmpfs) {
                                    setSource("");
                                }
                            }}
                            classNameContainer="col-span-2"
                            triggerClassName="w-full"
                        >
                            <SelectItem value={EMountType.Bind}>bind</SelectItem>
                            <SelectItem value={EMountType.Volume}>volume</SelectItem>
                            <SelectItem value={EMountType.Tmpfs}>tmpfs</SelectItem>
                            <SelectItem value={EMountType.Npipe}>npipe</SelectItem>
                            <SelectItem value={EMountType.Cluster}>cluster</SelectItem>
                            <SelectItem value={EMountType.Image}>image</SelectItem>
                        </SelectWithAddon>

                        {isComboboxSourceType(type) ? (
                            <ComboboxWithAddon<SourceOptionValue>
                                addonLeft="Source"
                                options={sourceOptions}
                                value={source || null}
                                onChange={(_, option) => {
                                    setSource(option?.name ?? "");
                                }}
                                placeholder="name"
                                emptyText="No sources"
                                valueKey="name"
                                classNameContainer="col-span-2"
                            />
                        ) : (
                            <InputWithAddOn
                                addonLeft="Source"
                                value={source}
                                onChange={event => {
                                    setSource(event.target.value);
                                }}
                                placeholder="source"
                                disabled={isTmpfs}
                                classNameContainer="col-span-2"
                            />
                        )}

                        <InputWithAddOn
                            addonLeft="Target"
                            value={target}
                            onChange={event => {
                                setTarget(event.target.value);
                            }}
                            placeholder="/path/in/container"
                            classNameContainer="col-span-2"
                        />

                        <SelectWithAddon
                            addonLeft="Consistency"
                            value={consistency}
                            onValueChange={value => {
                                setConsistency(value as EMountConsistency);
                            }}
                            classNameContainer="col-span-2"
                            triggerClassName="w-full"
                        >
                            <SelectItem value={EMountConsistency.Default}>default</SelectItem>
                            <SelectItem value={EMountConsistency.Consistent}>consistent</SelectItem>
                            <SelectItem value={EMountConsistency.Cached}>cached</SelectItem>
                            <SelectItem value={EMountConsistency.Delegated}>delegated</SelectItem>
                        </SelectWithAddon>
                    </div>

                    <Button
                        type="button"
                        variant="outline"
                        disabled={!canAdd}
                        onClick={() => {
                            if (!canAdd) {
                                toast.error("Please provide required mount fields");
                                return;
                            }

                            append({
                                type,
                                source: isTmpfs ? "" : source.trim(),
                                target: target.trim(),
                                consistency,
                            });

                            setSource("");
                            setTarget("");
                        }}
                    >
                        <Plus className="size-4" />
                        Add
                    </Button>

                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                            // setRefreshTick(prev => prev + 1);
                        }}
                    >
                        <RefreshCw className="size-4" />
                    </Button>

                    <a
                        href="/cluster/volumes"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-blue-500 hover:underline"
                    >
                        Manage volumes
                    </a>
                </div>

                <div className="divide-y divide-zinc-200">
                    {fields.map((field, index) => (
                        <div
                            key={field.id}
                            className="grid grid-cols-[1fr_1fr_1fr_1fr_auto] items-center gap-3 py-2"
                        >
                            <InputWithAddOn
                                addonLeft="Type"
                                value={field.type}
                                disabled
                                classNameContainer="w-full"
                            />
                            <InputWithAddOn
                                addonLeft="Source"
                                value={field.source}
                                disabled
                                classNameContainer="w-full"
                            />
                            <InputWithAddOn
                                addonLeft="Target"
                                value={field.target}
                                disabled
                                classNameContainer="w-full"
                            />
                            <InputWithAddOn
                                addonLeft="Consistency"
                                value={field.consistency}
                                disabled
                                classNameContainer="w-full"
                            />
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
