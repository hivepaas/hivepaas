import { useEffect, useState } from "react";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { Badge } from "@components/ui/badge";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { ChevronDown, ChevronRight, Plus, Trash2 } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";
import { EHttpPathMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

import { HttpConfigurableSections } from "./http-configurable-sections.com";

interface PathsSectionProps {
    domainIndex: number;
    readOnly?: boolean;
}

interface PathRowProps {
    domainIndex: number;
    pathIndex: number;
    expandedPaths: ReadonlySet<string>;
    onExpandedChange: (pathKey: string, open: boolean) => void;
    onRemove: () => void;
    readOnly?: boolean;
}

function PathRow({ domainIndex, pathIndex, expandedPaths, onExpandedChange, onRemove, readOnly = false }: PathRowProps) {
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
    const { field: enabled } = useController({ control, name: `${pathPrefix}.enabled` as never });
    const isEnabled = Boolean(enabled.value);

    const expanded = expandedPaths.has(path.value);

    return (
        <Collapsible
            open={expanded}
            onOpenChange={open => {
                onExpandedChange(path.value, open);
            }}
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
                    onConfirm={() => {
                        if (readOnly) {
                            return;
                        }

                        onRemove();
                    }}
                >
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        disabled={readOnly}
                    >
                        <Trash2 className="size-3.5" />
                    </Button>
                </PopConfirm>
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-3 pb-3 ml-3">
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
                        <>
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
                                    className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[pathError]} />
                            </InfoBlock>

                            <InfoBlock title="Match Mode">
                                <Tabs
                                    value={mode.value}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        mode.onChange(value);
                                    }}
                                    className="w-fit"
                                >
                                    <TabsList>
                                        <TabsTrigger
                                            value={EHttpPathMode.Exact}
                                            disabled={readOnly}
                                        >
                                            Exact
                                        </TabsTrigger>
                                        <TabsTrigger
                                            value={EHttpPathMode.Prefix}
                                            disabled={readOnly}
                                        >
                                            Prefix
                                        </TabsTrigger>
                                        <TabsTrigger
                                            value={EHttpPathMode.Regex}
                                            disabled={readOnly}
                                        >
                                            Regex
                                        </TabsTrigger>
                                    </TabsList>
                                </Tabs>
                            </InfoBlock>

                            <HttpConfigurableSections
                                basePath={basePath}
                                readOnly={readOnly}
                            />
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

export function PathsSection({ domainIndex, readOnly = false }: PathsSectionProps) {
    const [newPath, setNewPath] = useState("");
    const [newMode, setNewMode] = useState<EHttpPathMode>(EHttpPathMode.Prefix);
    const [expandedPaths, setExpandedPaths] = useState(() => new Set<string>());

    useEffect(() => {
        setExpandedPaths(new Set());
    }, [domainIndex]);

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
        if (readOnly) {
            return;
        }

        if (!newPath.trim()) return;
        append({ enabled: true, path: newPath.trim(), mode: newMode });
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
                                if (readOnly) {
                                    return;
                                }

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
                            disabled={readOnly}
                        />
                        <Tabs
                            value={newMode}
                            onValueChange={v => {
                                if (readOnly) {
                                    return;
                                }

                                setNewMode(v as EHttpPathMode);
                            }}
                            className="w-fit"
                        >
                            <TabsList>
                                <TabsTrigger
                                    value={EHttpPathMode.Exact}
                                    disabled={readOnly}
                                >
                                    Exact
                                </TabsTrigger>
                                <TabsTrigger
                                    value={EHttpPathMode.Prefix}
                                    disabled={readOnly}
                                >
                                    Prefix
                                </TabsTrigger>
                                <TabsTrigger
                                    value={EHttpPathMode.Regex}
                                    disabled={readOnly}
                                >
                                    Regex
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAdd}
                            disabled={readOnly || !newPath.trim()}
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
                        expandedPaths={expandedPaths}
                        onExpandedChange={(pathKey, open) => {
                            setExpandedPaths(prev => {
                                const next = new Set(prev);
                                if (open) {
                                    next.add(pathKey);
                                } else {
                                    next.delete(pathKey);
                                }
                                return next;
                            });
                        }}
                        onRemove={() => {
                            if (readOnly) {
                                return;
                            }

                            remove(index);
                        }}
                        readOnly={readOnly}
                    />
                ))}
            </div>
        </>
    );
}
