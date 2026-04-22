import { useState } from "react";

import { Button, FieldError, Input } from "@components/ui";
import { Badge } from "@components/ui/badge";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { ChevronDown, ChevronRight, Plus, Trash2 } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";
import { EHttpPathMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

import { HttpConfigurableSections } from "./http-configurable-sections.com";

interface PathsSectionProps {
    domainIndex: number;
}

interface PathRowProps {
    domainIndex: number;
    pathIndex: number;
    onRemove: () => void;
}

function PathRow({ domainIndex, pathIndex, onRemove }: PathRowProps) {
    const [expanded, setExpanded] = useState(false);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const pathPrefix = `domains.${domainIndex}.paths.${pathIndex}`;
    const basePath = pathPrefix;

    const {
        field: path,
        fieldState: { error: pathError },
    } = useController({ control, name: `${pathPrefix}.path` as never });

    const { field: mode } = useController({ control, name: `${pathPrefix}.mode` as never });

    return (
        <Collapsible
            open={expanded}
            onOpenChange={setExpanded}
        >
            <div className="flex items-center gap-2 rounded-md border p2 bg-accent">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex items-center gap-2 flex-1 text-left"
                    >
                        {expanded ? (
                            <ChevronDown className="size-4 shrink-0" />
                        ) : (
                            <ChevronRight className="size-4 shrink-0" />
                        )}
                        <span className="font-mono text-sm text-red-500">Path: {path.value}</span>
                        <Badge>{mode.value}</Badge>
                    </button>
                </CollapsibleTrigger>
                <PopConfirm
                    title="Remove path"
                    description="Confirm deletion of this path?"
                    confirmText="Remove"
                    cancelText="Cancel"
                    variant="destructive"
                    onConfirm={onRemove}
                >
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                    >
                        <Trash2 className="size-3.5" />
                    </Button>
                </PopConfirm>
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-3 pb-3 ml-3">
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Path"
                                isRequired
                            />
                        }
                    >
                        <Input
                            {...path}
                            value={path.value}
                            onChange={path.onChange}
                            placeholder="/api/v1"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[pathError]} />
                    </InfoBlock>

                    <InfoBlock title="Match Mode">
                        <Tabs
                            value={mode.value}
                            onValueChange={mode.onChange}
                            className="w-fit"
                        >
                            <TabsList>
                                <TabsTrigger value={EHttpPathMode.Exact}>Exact</TabsTrigger>
                                <TabsTrigger value={EHttpPathMode.Prefix}>Prefix</TabsTrigger>
                                <TabsTrigger value={EHttpPathMode.Regex}>Regex</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <HttpConfigurableSections basePath={basePath} />
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

export function PathsSection({ domainIndex }: PathsSectionProps) {
    const [newPath, setNewPath] = useState("");
    const [newMode, setNewMode] = useState<EHttpPathMode>(EHttpPathMode.Prefix);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const { fields, append, remove } = useFieldArray({
        control,
        name: `domains.${domainIndex}.paths`,
    });

    function handleAdd() {
        if (!newPath.trim()) return;
        append({ path: newPath.trim(), mode: newMode });
        setNewPath("");
        setNewMode(EHttpPathMode.Prefix);
    }

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Path"
                        content="Override settings for specific request paths."
                    />
                }
            >
                <div className="flex flex-col gap-3 max-w-[700px]">
                    <div className="flex items-center gap-2">
                        <Input
                            value={newPath}
                            onChange={e => {
                                setNewPath(e.target.value);
                            }}
                            placeholder="/api/admin"
                            className="flex-1"
                            onKeyDown={e => {
                                if (e.key === "Enter") {
                                    e.preventDefault();
                                    handleAdd();
                                }
                            }}
                        />
                        <Tabs
                            value={newMode}
                            onValueChange={v => {
                                setNewMode(v as EHttpPathMode);
                            }}
                            className="w-fit"
                        >
                            <TabsList>
                                <TabsTrigger value={EHttpPathMode.Exact}>Exact</TabsTrigger>
                                <TabsTrigger value={EHttpPathMode.Prefix}>Prefix</TabsTrigger>
                                <TabsTrigger value={EHttpPathMode.Regex}>Regex</TabsTrigger>
                            </TabsList>
                        </Tabs>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAdd}
                            disabled={!newPath.trim()}
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>
                </div>
            </InfoBlock>
            <div className="flex flex-col gap-2">
                {fields.map((field, index) => (
                    <PathRow
                        key={field.id}
                        domainIndex={domainIndex}
                        pathIndex={index}
                        onRemove={() => {
                            remove(index);
                        }}
                    />
                ))}
            </div>
        </>
    );
}
