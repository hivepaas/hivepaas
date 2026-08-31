import { useMemo, useState } from "react";

import { useFieldArray, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectNetworksQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { FieldListLayout } from "@application/shared/form";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

const networkAttachmentsGridClass = "grid flex-1 min-w-0 w-full grid-cols-2 gap-2 items-center [&>*]:min-w-0";

type NetworkOptionValue = {
    id: string;
    name: string;
};

export function NetworksFields({ readOnly = false }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const { control, getValues } = useFormContext<
        AppConfigNetworksFormSchemaInput,
        unknown,
        AppConfigNetworksFormSchemaOutput
    >();
    const { fields, append, remove, update } = useFieldArray({
        control,
        name: "networkAttachments",
    });

    const [search, setSearch] = useState("");
    const [selectedNetwork, setSelectedNetwork] = useState<NetworkOptionValue | null>(null);
    const [aliasesText, setAliasesText] = useState("");
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [draftNetwork, setDraftNetwork] = useState<NetworkOptionValue | null>(null);
    const [draftAliases, setDraftAliases] = useState("");

    const {
        data: { data: projectNetworks } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectNetworksQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        search,
    });

    const comboboxOptions = useMemo(() => {
        return projectNetworks.map(n => ({
            value: { id: n.id, name: n.name },
            label: n.name,
        }));
    }, [projectNetworks]);

    const handleAdd = () => {
        if (readOnly) {
            return;
        }

        if (!selectedNetwork) {
            toast.error("Please select a network");
            return;
        }

        append({ id: selectedNetwork.id, name: selectedNetwork.name, aliasesText: aliasesText.trim() });
        setSelectedNetwork(null);
        setAliasesText("");
        setEditingIndex(null);
        setDraftNetwork(null);
        setDraftAliases("");
    };

    const handleStartEdit = (index: number) => {
        if (readOnly) {
            return;
        }

        // useFieldArray shadows the form `id` with its own React key, so read from form values.
        const currentAttachment = getValues(`networkAttachments.${index}`);
        setEditingIndex(index);
        setDraftNetwork({ id: currentAttachment.id, name: currentAttachment.name });
        setDraftAliases(currentAttachment.aliasesText);
    };

    const handleSaveEdit = (index: number) => {
        if (readOnly || editingIndex !== index) {
            return;
        }

        if (!draftNetwork) {
            toast.error("Please select a network");
            return;
        }

        update(index, {
            id: draftNetwork.id,
            name: draftNetwork.name,
            aliasesText: draftAliases.trim(),
        });
        setEditingIndex(null);
        setDraftNetwork(null);
        setDraftAliases("");
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftNetwork(null);
        setDraftAliases("");
    };

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Networks"
                    content="Attach app to networks and optionally define network aliases."
                />
            }
        >
            <FieldListLayout
                className="max-w-[590px]"
                inputsClassName={networkAttachmentsGridClass}
                inputRow={
                    <>
                        <Combobox<NetworkOptionValue>
                            options={comboboxOptions}
                            value={selectedNetwork?.id ?? null}
                            onChange={(_, option) => {
                                if (readOnly) {
                                    return;
                                }

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
                            disabled={readOnly}
                        />
                        <InputWithAddOn
                            addonLeft="Alias"
                            value={aliasesText}
                            onChange={e => {
                                if (readOnly) {
                                    return;
                                }

                                setAliasesText(e.target.value);
                            }}
                            placeholder="alias1 alias2"
                            disabled={readOnly}
                        />
                    </>
                }
                onAdd={handleAdd}
                disabled={readOnly}
                items={fields.map((field, index) => {
                    const isEditing = !readOnly && editingIndex === index;

                    return {
                        id: field.id,
                        content: (
                            <div className={networkAttachmentsGridClass}>
                                {isEditing ? (
                                    <>
                                        <Combobox<NetworkOptionValue>
                                            options={comboboxOptions}
                                            value={draftNetwork?.id ?? null}
                                            onChange={(_, option) => {
                                                setDraftNetwork(option ?? null);
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
                                            value={draftAliases}
                                            onChange={e => {
                                                setDraftAliases(e.target.value);
                                            }}
                                            onKeyDown={e => {
                                                if (e.key === "Enter") {
                                                    e.preventDefault();
                                                    handleSaveEdit(index);
                                                }
                                                if (e.key === "Escape") {
                                                    e.preventDefault();
                                                    handleCancelEdit();
                                                }
                                            }}
                                            placeholder="alias1 alias2"
                                        />
                                    </>
                                ) : (
                                    <>
                                        <span className="text-sm wrap-break-word min-w-0">
                                            {field.name || field.id}
                                        </span>
                                        <span className="text-sm wrap-break-word min-w-0">{field.aliasesText}</span>
                                    </>
                                )}
                            </div>
                        ),
                        isEditing,
                        onEdit: readOnly
                            ? undefined
                            : () => {
                                  if (isEditing) {
                                      handleSaveEdit(index);
                                      return;
                                  }

                                  handleStartEdit(index);
                              },
                        onRemove: () => {
                            if (editingIndex === index) {
                                handleCancelEdit();
                            } else if (editingIndex !== null && editingIndex > index) {
                                setEditingIndex(editingIndex - 1);
                            }

                            remove(index);
                        },
                    };
                })}
            />
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
