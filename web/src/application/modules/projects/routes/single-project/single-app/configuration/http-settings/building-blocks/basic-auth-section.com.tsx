import { useMemo, useState } from "react";

import { Button, FieldError } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { ProjectBasicAuthQueries } from "@application/modules/projects/data";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface BasicAuthSectionProps {
    prefix: string;
    onRemove: () => void;
}

export function BasicAuthSection({ prefix, onRemove }: BasicAuthSectionProps) {
    const [open, setOpen] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const {
        field: nameField,
        fieldState: { error: nameError },
    } = useController({
        control,
        name: `${prefix}.name` as never,
    });

    const { field: idField } = useController({
        control,
        name: `${prefix}.id` as never,
    });
    const basicAuthValue = useWatch({
        control,
        name: prefix as never,
    }) as { id?: string; name?: string } | undefined;

    const {
        data: { data: basicAuths } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectBasicAuthQueries.useFindManyPaginated({
        projectID: projectId,
        search: searchQuery,
    });

    const comboboxOptions = useMemo(() => {
        return basicAuths.map(item => ({
            value: {
                id: item.id,
                name: item.name,
            },
            label: item.name,
        }));
    }, [basicAuths]);

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <div className="flex w-full items-center gap-1 rounded-md border border-dashed px-1">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex min-w-0 flex-1 items-center gap-2 px-2 py-2 text-sm font-medium hover:bg-accent"
                    >
                        {open ? (
                            <ChevronDown className="size-4 shrink-0" />
                        ) : (
                            <ChevronRight className="size-4 shrink-0" />
                        )}
                        Basic Auth
                    </button>
                </CollapsibleTrigger>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="shrink-0 text-muted-foreground hover:text-destructive"
                    title="Remove section"
                    onClick={onRemove}
                >
                    <X className="size-4" />
                </Button>
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Credential">
                        <Combobox
                            options={comboboxOptions}
                            value={basicAuthValue?.id ?? null}
                            onChange={(_, option) => {
                                if (!option) {
                                    idField.onChange("");
                                    nameField.onChange("");
                                    return;
                                }

                                idField.onChange(option.id);
                                nameField.onChange(option.name);
                            }}
                            onSearch={setSearchQuery}
                            placeholder="Select basic auth credential"
                            searchable
                            closeOnSelect
                            emptyText="No basic auth credentials available"
                            className="max-w-[400px]"
                            valueKey="id"
                            loading={isFetching}
                            onRefresh={() => void refetch()}
                            isRefreshing={isRefetching}
                            allowClear
                        />
                        <FieldError errors={[nameError]} />
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
