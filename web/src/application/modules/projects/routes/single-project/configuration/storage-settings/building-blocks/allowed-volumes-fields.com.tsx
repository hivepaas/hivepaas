import { useMemo, useState } from "react";

import { Button, FieldError } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { type FieldPath, useFieldArray, useFormContext, useFormState } from "react-hook-form";
import { toast } from "sonner";
import { ClusterVolumesQueries } from "~/cluster/data";

import { ComboboxWithAddon, InfoBlock, LabelWithInfo, PopConfirm } from "@application/shared/components";

import type { ProjectStorageSettingsFormSchemaInput, ProjectStorageSettingsFormSchemaOutput } from "../schemas";

import {
    STORAGE_SETTINGS_CONTROL_CLASS_NAME,
    STORAGE_SETTINGS_FIELD_TITLE_WIDTH,
} from "./storage-settings-layout.constants";

type VolumeFieldPath = Extract<
    FieldPath<ProjectStorageSettingsFormSchemaInput>,
    "volumeSettings.volumes" | "clusterVolumeSettings.volumes"
>;

type VolumeOption = {
    id: string;
    name: string;
};

export function AllowedVolumesFields({ name, type, readOnly = false }: Props) {
    const [search, setSearch] = useState("");
    const [selectedVolume, setSelectedVolume] = useState<VolumeOption | null>(null);
    const { control } = useFormContext<
        ProjectStorageSettingsFormSchemaInput,
        unknown,
        ProjectStorageSettingsFormSchemaOutput
    >();
    const { fields, append, remove } = useFieldArray({
        control,
        name,
        keyName: "fieldId",
    });
    const { errors } = useFormState<ProjectStorageSettingsFormSchemaInput>();
    const volumesQuery = ClusterVolumesQueries.useList({
        type,
        search,
        pagination: {
            page: 1,
            size: 20,
        },
    });

    const options = useMemo(
        () =>
            (volumesQuery.data?.data ?? []).map(volume => ({
                label: volume.name,
                value: {
                    id: volume.id,
                    name: volume.name,
                },
            })),
        [volumesQuery.data],
    );

    const fieldError =
        name === "volumeSettings.volumes" ? errors.volumeSettings?.volumes : errors.clusterVolumeSettings?.volumes;

    function handleAdd() {
        if (readOnly || !selectedVolume) {
            return;
        }

        const exists = fields.some(field => field.id === selectedVolume.id);

        if (exists) {
            toast.error(`"${selectedVolume.name}" already exists`);
            return;
        }

        append(selectedVolume);
        setSelectedVolume(null);
    }

    return (
        <InfoBlock
            titleWidth={STORAGE_SETTINGS_FIELD_TITLE_WIDTH}
            title={
                <LabelWithInfo
                    label="Allowed Volumes"
                    content="Volumes that apps in this project are allowed to mount."
                />
            }
        >
            <div className={`flex flex-col gap-2 ${STORAGE_SETTINGS_CONTROL_CLASS_NAME}`}>
                <div className="flex gap-2">
                    <ComboboxWithAddon<VolumeOption>
                        addonLeft="Volume"
                        value={selectedVolume?.id}
                        onChange={(_, option) => {
                            setSelectedVolume(option);
                        }}
                        onSearch={setSearch}
                        onRefresh={() => void volumesQuery.refetch()}
                        isRefreshing={volumesQuery.isRefetching}
                        loading={volumesQuery.isLoading}
                        valueKey="id"
                        options={options}
                        placeholder="volume"
                        classNameContainer="max-w-[460px]"
                        disabled={readOnly}
                    />

                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleAdd}
                        disabled={readOnly || !selectedVolume}
                    >
                        <Plus className="size-4" /> Add
                    </Button>
                </div>

                {fields.length > 0 && (
                    <div className="flex flex-col divide-y">
                        {fields.map((field, index) => (
                            <div
                                key={field.fieldId}
                                className="flex items-center justify-between gap-2 py-1.5"
                            >
                                <div className="flex-1 grid grid-cols-1 items-center">
                                    <span className="break-words text-sm">{field.name}</span>
                                </div>
                                <div className="w-[76px]">
                                    <PopConfirm
                                        title="Remove volume"
                                        variant="destructive"
                                        confirmText="Remove"
                                        cancelText="Cancel"
                                        description="Are you sure you want to remove this volume?"
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

                <FieldError errors={[fieldError]} />
            </div>
        </InfoBlock>
    );
}

type Props = {
    name: VolumeFieldPath;
    type: "cluster" | "volume";
    readOnly?: boolean;
};
