import { useMemo, useState } from "react";

import { Avatar, Button, FieldError } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext, useFormState } from "react-hook-form";
import { toast } from "sonner";
import { ProjectAppsQueries } from "~/projects/data";

import { ComboboxWithAddon, InfoBlock, PopConfirm } from "@application/shared/components";

import type { AppFeatureSettingsFormSchemaInput, AppFeatureSettingsFormSchemaOutput } from "../schemas";

type AppOption = {
    id: string;
    name: string;
    photo?: string;
};

export function DbAppsToCloneFields({ projectID, env, appID, readOnly = false }: Props) {
    const [search, setSearch] = useState("");
    const [selectedApp, setSelectedApp] = useState<AppOption | null>(null);
    const { control } = useFormContext<
        AppFeatureSettingsFormSchemaInput,
        unknown,
        AppFeatureSettingsFormSchemaOutput
    >();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "previewSettings.appsToClone",
        keyName: "fieldId",
    });
    const { errors } = useFormState<AppFeatureSettingsFormSchemaInput>();

    const appsQuery = ProjectAppsQueries.useFindManyPaginated({
        projectID,
        env,
        search,
        pagination: {
            page: 1,
            size: 20,
        },
    });

    const selectedIds = useMemo(() => new Set(fields.map(field => field.id)), [fields]);

    const options = useMemo(
        () =>
            (appsQuery.data?.data ?? [])
                .filter(app => app.id !== appID && !selectedIds.has(app.id))
                .map(app => ({
                    label: app.name,
                    value: {
                        id: app.id,
                        name: app.name,
                        photo: app.photo,
                    },
                })),
        [appsQuery.data, appID, selectedIds],
    );

    const fieldError = errors.previewSettings?.appsToClone;

    function handleAdd() {
        if (readOnly || !selectedApp) {
            return;
        }

        const exists = fields.some(field => field.id === selectedApp.id);

        if (exists) {
            toast.error(`"${selectedApp.name}" already exists`);
            return;
        }

        append(selectedApp);
        setSelectedApp(null);
    }

    return (
        <InfoBlock title="DB Apps to Clone">
            <div className="flex max-w-[545px] flex-col gap-2">
                <div className="flex gap-2">
                    <ComboboxWithAddon<AppOption>
                        addonLeft="App"
                        value={selectedApp?.id}
                        onChange={(_, option) => {
                            setSelectedApp(option);
                        }}
                        onSearch={setSearch}
                        onRefresh={() => void appsQuery.refetch()}
                        isRefreshing={appsQuery.isRefetching}
                        loading={appsQuery.isLoading}
                        valueKey="id"
                        options={options}
                        placeholder="select app to add"
                        classNameContainer="max-w-[460px]"
                        disabled={readOnly}
                    />

                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleAdd}
                        disabled={readOnly || !selectedApp}
                    >
                        <Plus className="size-4" /> Add
                    </Button>
                </div>

                {fields.length > 0 && (
                    <div className="flex w-full flex-col divide-y">
                        {fields.map((field, index) => (
                            <div
                                key={field.fieldId}
                                className="grid grid-cols-[minmax(0,1fr)_76px] items-center gap-2 py-1.5"
                            >
                                <div className="flex min-w-0 items-center gap-2">
                                    <Avatar
                                        name={field.name}
                                        src={field.photo}
                                    />
                                    <span className="truncate text-sm">{field.name}</span>
                                </div>
                                <div className="w-[76px]">
                                    <PopConfirm
                                        title="Remove app"
                                        variant="destructive"
                                        confirmText="Remove"
                                        cancelText="Cancel"
                                        description="Are you sure you want to remove this app?"
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
    projectID: string;
    env: string;
    appID: string;
    readOnly?: boolean;
};
