import { useMemo, useState } from "react";

import { Input } from "@components/ui";
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
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

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
    const [draftAliases, setDraftAliases] = useState("");

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
        setDraftAliases("");
    };

    const handleStartEdit = (index: number, currentAliases: string) => {
        if (readOnly) {
            return;
        }

        setEditingIndex(index);
        setDraftAliases(currentAliases);
    };

    const handleSaveEdit = (index: number) => {
        if (readOnly || editingIndex !== index) {
            return;
        }

        // useFieldArray shadows the form `id` with its own React key, so read from form values.
        const currentAttachment = getValues(`networkAttachments.${index}`);
        update(index, {
            id: currentAttachment.id,
            name: currentAttachment.name,
            aliasesText: draftAliases.trim(),
        });
        setEditingIndex(null);
        setDraftAliases("");
    };

    const handleCancelEdit = () => {
        setEditingIndex(null);
        setDraftAliases("");
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
                                <span className="text-sm wrap-break-word min-w-0">{field.name || field.id}</span>
                                {isEditing ? (
                                    <Input
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
                                        className="h-8"
                                    />
                                ) : (
                                    <span className="text-sm wrap-break-word min-w-0">{field.aliasesText}</span>
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

                                  handleStartEdit(index, field.aliasesText);
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
