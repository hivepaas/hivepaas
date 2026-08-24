import { useMemo, useState } from "react";

import { Button, FieldError } from "@components/ui";
import { Plus, Terminal, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext, useFormState } from "react-hook-form";
import { toast } from "sonner";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";

import { AppLink, ComboboxWithAddon, InfoBlock, LabelWithInfo, PopConfirm } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import {
    type AppFeatureSettingsFormSchemaInput,
    type AppFeatureSettingsFormSchemaOutput,
    FEATURE_SETTINGS_TITLE_WIDTH,
} from "../schemas";

type CommandOption = {
    id: string;
    name: string;
};

export function PreviewCommandsFields({ projectID, readOnly = false }: Props) {
    const [search, setSearch] = useState("");
    const [selectedCommand, setSelectedCommand] = useState<CommandOption | null>(null);
    const { control } = useFormContext<
        AppFeatureSettingsFormSchemaInput,
        unknown,
        AppFeatureSettingsFormSchemaOutput
    >();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "previewSettings.commands",
        keyName: "fieldId",
    });
    const { errors } = useFormState<AppFeatureSettingsFormSchemaInput>();

    const commandsQuery = ProjectCommandTemplateQueries.useFindManyPaginated(
        {
            projectID,
            search,
            pagination: {
                page: 1,
                size: 50,
            },
        },
        {
            enabled: Boolean(projectID),
        },
    );

    const selectedIds = useMemo(() => new Set(fields.map(field => field.id)), [fields]);

    const options = useMemo(
        () =>
            (commandsQuery.data?.data ?? [])
                .filter(cmd => !selectedIds.has(cmd.id))
                .map(cmd => ({
                    label: cmd.name,
                    value: {
                        id: cmd.id,
                        name: cmd.name,
                    },
                })),
        [commandsQuery.data, selectedIds],
    );

    const fieldError = errors.previewSettings?.commands;

    function handleAdd() {
        if (readOnly || !selectedCommand) {
            return;
        }

        const exists = fields.some(field => field.id === selectedCommand.id);

        if (exists) {
            toast.error(`"${selectedCommand.name}" already exists`);
            return;
        }

        append(selectedCommand);
        setSelectedCommand(null);
    }

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Commands to Run on Preview Creation"
                    content="Command templates to execute inside the current app container when creating a preview (e.g. create remote DB branch, run seed scripts)."
                />
            }
            titleWidth={FEATURE_SETTINGS_TITLE_WIDTH}
        >
            <div className="flex max-w-[545px] flex-col gap-2">
                <div className="flex gap-2">
                    <ComboboxWithAddon<CommandOption>
                        addonLeft="Command"
                        value={selectedCommand?.id}
                        onChange={(_, option) => {
                            setSelectedCommand(option);
                        }}
                        onSearch={setSearch}
                        onRefresh={() => void commandsQuery.refetch()}
                        isRefreshing={commandsQuery.isRefetching}
                        loading={commandsQuery.isLoading}
                        valueKey="id"
                        options={options}
                        placeholder="select command template to add"
                        classNameContainer="max-w-[460px]"
                        disabled={readOnly}
                    />

                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleAdd}
                        disabled={readOnly || !selectedCommand}
                    >
                        <Plus className="size-4" /> Add
                    </Button>
                </div>

                {fields.length > 0 && (
                    <div className="flex w-full flex-col divide-y">
                        {fields.map((field, index) => (
                            <div
                                key={field.fieldId}
                                className="grid grid-cols-[minmax(0,1fr)_auto_76px] items-center gap-2 py-1.5"
                            >
                                <div className="flex min-w-0 items-center gap-2">
                                    <Terminal className="size-4 shrink-0 text-muted-foreground" />
                                    <span className="truncate text-sm font-medium">{field.name}</span>
                                </div>
                                <AppLink.Basic
                                    to={ROUTE.projects.single.providerConfiguration.commandTemplates.edit.$route(
                                        projectID,
                                        field.id,
                                    )}
                                    className="shrink-0 text-xs text-link"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    ignorePrevPath
                                >
                                    View
                                </AppLink.Basic>
                                <div className="flex w-[76px] justify-end">
                                    <PopConfirm
                                        title="Remove command"
                                        variant="destructive"
                                        confirmText="Remove"
                                        cancelText="Cancel"
                                        description="Are you sure you want to remove this command template?"
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
    readOnly?: boolean;
};
