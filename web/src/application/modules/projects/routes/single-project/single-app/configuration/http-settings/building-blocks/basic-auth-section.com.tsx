import { useMemo, useState } from "react";

import { Button, Checkbox, Field, FieldError } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppLink, Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import { ProjectBasicAuthQueries } from "@application/modules/projects/data";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface BasicAuthSectionProps {
    prefix: string;
    readOnly?: boolean;
    onRemove: () => void;
}

export function BasicAuthSection({ prefix, readOnly = false, onRemove }: BasicAuthSectionProps) {
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
    const { field: enabled } = useController({
        control,
        name: `${prefix}.enabled` as never,
    });
    const isEnabled = Boolean(enabled.value);
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
            <div className="flex justify-between items-center font-medium bg-accent py-2 px-3 rounded-lg">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex min-w-0 flex-1 items-center gap-2 text-sm font-medium hover:bg-accent"
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
                    className="shrink-0 text-muted-foreground hover:text-destructive h-fit"
                    title="Remove section"
                    onClick={() => {
                        if (readOnly) {
                            return;
                        }

                        onRemove();
                    }}
                    disabled={readOnly}
                >
                    <X className="size-4" />
                </Button>
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Enabled">
                        <Checkbox
                            checked={isEnabled}
                            onCheckedChange={value => {
                                if (readOnly) {
                                    return;
                                }

                                enabled.onChange(value);
                            }}
                            disabled={readOnly}
                        />
                    </InfoBlock>
                    {isEnabled && (
                        <InfoBlock title="Basic Auth">
                            <Field className="">
                                <Combobox
                                    options={comboboxOptions}
                                    value={basicAuthValue?.id ?? null}
                                    onChange={(_, option) => {
                                        if (readOnly) {
                                            return;
                                        }

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
                                    className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                    valueKey="id"
                                    loading={isFetching}
                                    onRefresh={() => void refetch()}
                                    isRefreshing={isRefetching}
                                    allowClear
                                    disabled={readOnly}
                                />
                                <FieldError errors={[nameError]} />
                                <div className="text-xs">
                                    <AppLink.Basic
                                        to={ROUTE.projects.single.providerConfiguration.basicAuth.$route(projectId)}
                                        className="text-link"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        Configure Basic Auth Items
                                    </AppLink.Basic>
                                </div>
                            </Field>
                        </InfoBlock>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
