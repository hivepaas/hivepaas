import { useMemo, useState } from "react";

import { Button } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { Plus, Trash2 } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";
import { toast } from "sonner";
import { NodesQueries } from "~/cluster/data";

import {
    ComboboxWithAddon,
    InfoBlock,
    InputWithAddOn,
    LabelWithInfo,
    PopConfirm,
} from "@application/shared/components";

import type { SettingsImageBuildFormSchemaInput, SettingsImageBuildFormSchemaOutput } from "../schemas";

type NodeOption = {
    id: string;
    name: string;
};

function formatBuildNodeLabel(node: NodeOption): string {
    return `${node.id} (name: ${node.name})`;
}

export function BuildWorkerFields({ readOnly = false }: { readOnly?: boolean }) {
    const [nodeSearch, setNodeSearch] = useState("");
    const [selectedNode, setSelectedNode] = useState<NodeOption | null>(null);
    const [labelInput, setLabelInput] = useState("");
    const { control } = useFormContext<
        SettingsImageBuildFormSchemaInput,
        unknown,
        SettingsImageBuildFormSchemaOutput
    >();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "workers.nodes",
        keyName: "fieldId",
    });
    const { field: labelsField } = useController({ control, name: "workers.nodeLabels" });
    const { field: maxParallelismField } = useController({ control, name: "workers.maxParallelism" });

    const nodesQuery = NodesQueries.useFindManyPaginated({
        search: nodeSearch,
        pagination: {
            page: 1,
            size: 50,
        },
    });

    const selectedIds = useMemo(() => new Set(fields.map(field => field.id)), [fields]);

    const nodeOptions = useMemo(
        () =>
            (nodesQuery.data?.data ?? [])
                .filter(node => node.refId && !selectedIds.has(node.refId))
                .map(node => {
                    const name = node.name || node.hostname || node.refId;

                    return {
                        label: formatBuildNodeLabel({ id: node.refId, name }),
                        value: {
                            id: node.refId,
                            name,
                        },
                    };
                }),
        [nodesQuery.data, selectedIds],
    );

    const nodeLabels = labelsField.value;

    function handleAddNode() {
        if (readOnly || !selectedNode) {
            return;
        }

        if (fields.some(field => field.id === selectedNode.id)) {
            toast.error(`"${selectedNode.id}" already exists`);
            return;
        }

        append(selectedNode);
        setSelectedNode(null);
    }

    function handleAddLabel() {
        if (readOnly) {
            return;
        }

        const trimmed = labelInput.trim();
        if (!trimmed) {
            return;
        }

        if (nodeLabels.includes(trimmed)) {
            toast.error(`"${trimmed}" already exists`);
            return;
        }

        labelsField.onChange([...nodeLabels, trimmed]);
        setLabelInput("");
    }

    function handleRemoveLabel(index: number) {
        if (readOnly) {
            return;
        }

        labelsField.onChange(nodeLabels.filter((_, labelIndex) => labelIndex !== index));
    }

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={220}
                title="Build Node IDs"
            >
                <div className="flex flex-col gap-2">
                    <div className="flex gap-2">
                        <ComboboxWithAddon<NodeOption>
                            addonLeft="ID"
                            value={selectedNode?.id}
                            onChange={(_, option) => {
                                setSelectedNode(option);
                            }}
                            onSearch={setNodeSearch}
                            onRefresh={() => void nodesQuery.refetch()}
                            isRefreshing={nodesQuery.isRefetching}
                            loading={nodesQuery.isLoading}
                            valueKey="id"
                            options={nodeOptions}
                            placeholder="<node-id>"
                            classNameContainer="max-w-[460px]"
                            disabled={readOnly}
                        />
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAddNode}
                            disabled={readOnly || !selectedNode}
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>

                    {fields.length > 0 && (
                        <div className="flex w-full max-w-[545px] flex-col divide-y">
                            {fields.map((field, index) => (
                                <div
                                    key={field.fieldId}
                                    className="grid grid-cols-[minmax(0,1fr)_76px] items-center gap-2 py-1.5"
                                >
                                    <span className="break-words text-sm">{formatBuildNodeLabel(field)}</span>
                                    <div className="w-[76px]">
                                        <PopConfirm
                                            title="Remove node"
                                            variant="destructive"
                                            confirmText="Remove"
                                            cancelText="Cancel"
                                            description="Are you sure you want to remove this node?"
                                            onConfirm={() => {
                                                if (!readOnly) {
                                                    remove(index);
                                                }
                                            }}
                                        >
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 rounded-md text-zinc-400 hover:bg-red-50 hover:text-red-500"
                                                disabled={readOnly}
                                            >
                                                <Trash2 className="size-3.5" />
                                            </Button>
                                        </PopConfirm>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title="Build Node Labels"
            >
                <div className="flex flex-col gap-2">
                    <div className="flex gap-2">
                        <InputWithAddOn
                            addonLeft="Label"
                            value={labelInput}
                            onChange={event => {
                                if (!readOnly) {
                                    setLabelInput(event.target.value);
                                }
                            }}
                            placeholder="build-node"
                            classNameContainer="max-w-[460px]"
                            disabled={readOnly}
                            onKeyDown={event => {
                                if (event.key === "Enter") {
                                    event.preventDefault();
                                    handleAddLabel();
                                }
                            }}
                        />
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAddLabel}
                            disabled={readOnly || labelInput.trim() === ""}
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>

                    {nodeLabels.length > 0 && (
                        <div className="flex w-full max-w-[545px] flex-col divide-y">
                            {nodeLabels.map((label, index) => (
                                <div
                                    key={`${label}-${index}`}
                                    className="grid grid-cols-[minmax(0,1fr)_76px] items-center gap-2 py-1.5"
                                >
                                    <span className="break-words text-sm">{label}</span>
                                    <div className="w-[76px]">
                                        <PopConfirm
                                            title="Remove label"
                                            variant="destructive"
                                            confirmText="Remove"
                                            cancelText="Cancel"
                                            description="Are you sure you want to remove this label?"
                                            onConfirm={() => {
                                                handleRemoveLabel(index);
                                            }}
                                        >
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 rounded-md text-zinc-400 hover:bg-red-50 hover:text-red-500"
                                                disabled={readOnly}
                                            >
                                                <Trash2 className="size-3.5" />
                                            </Button>
                                        </PopConfirm>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Max Parallelism"
                        content="Maximum number of image builds that can run concurrently on each node. Set to 0 for unlimited."
                    />
                }
            >
                <InputNumber
                    value={maxParallelismField.value}
                    onValueChange={value => {
                        maxParallelismField.onChange(value ?? 0);
                    }}
                    className="max-w-[100px]"
                    min={0}
                    decimalScale={0}
                    fixedDecimalScale={false}
                    disabled={readOnly}
                />
            </InfoBlock>
        </div>
    );
}
