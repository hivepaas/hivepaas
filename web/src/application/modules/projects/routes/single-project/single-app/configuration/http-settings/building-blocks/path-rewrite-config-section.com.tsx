import { useEffect, useRef, useState } from "react";

import { Button, Checkbox, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext, useWatch } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

type PathRewriteMode = "addPrefix" | "stripPrefix" | "replace";

function deriveMode(values: { prefixAdd?: string; pathReplace?: string; prefixStrip?: string }): PathRewriteMode {
    if (values.pathReplace?.trim()) {
        return "replace";
    }
    if (values.prefixStrip?.trim()) {
        return "stripPrefix";
    }
    if (values.prefixAdd?.trim()) {
        return "addPrefix";
    }
    return "addPrefix";
}

interface PathRewriteConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    readOnly?: boolean;
    onRemove?: () => void;
}

export function PathRewriteConfigSection({
    prefix,
    autoExpandToken,
    readOnly = false,
    onRemove,
}: PathRewriteConfigSectionProps) {
    const [open, setOpen] = useState(false);
    const [mode, setMode] = useState<PathRewriteMode>("addPrefix");
    const initializedModeForPrefixRef = useRef<string | null>(null);

    const { control, setValue } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const setFormValue = setValue as (name: string, value: unknown, opts?: object) => void;

    const pathRewriteValues = useWatch({
        control,
        name: prefix as never,
    }) as
        | {
              prefixAdd?: string;
              prefixStrip?: string;
              pathReplace?: string;
          }
        | undefined;

    useEffect(() => {
        if (autoExpandToken !== undefined) {
            setOpen(true);
        }
    }, [autoExpandToken]);

    useEffect(() => {
        if (initializedModeForPrefixRef.current === prefix || pathRewriteValues == null) {
            return;
        }

        setMode(deriveMode(pathRewriteValues));
        initializedModeForPrefixRef.current = prefix;
    }, [prefix, pathRewriteValues]);

    const { field: enabled } = useController({ control, name: `${prefix}.enabled` as never });
    const { field: prefixAdd } = useController({ control, name: `${prefix}.prefixAdd` as never });
    const { field: prefixStrip } = useController({ control, name: `${prefix}.prefixStrip` as never });
    const { field: prefixStripIsRegex } = useController({ control, name: `${prefix}.prefixStripIsRegex` as never });
    const { field: pathReplace } = useController({ control, name: `${prefix}.pathReplace` as never });
    const { field: pathReplaceIsRegex } = useController({ control, name: `${prefix}.pathReplaceIsRegex` as never });
    const { field: pathReplaceWith } = useController({ control, name: `${prefix}.pathReplaceWith` as never });
    const isEnabled = Boolean(enabled.value);

    function handleModeChange(nextMode: PathRewriteMode) {
        if (readOnly) {
            return;
        }

        setMode(nextMode);
        const opts = { shouldDirty: true, shouldValidate: true };

        if (nextMode === "addPrefix") {
            setFormValue(`${prefix}.prefixStrip`, "", opts);
            setFormValue(`${prefix}.prefixStripIsRegex`, false, opts);
            setFormValue(`${prefix}.pathReplace`, "", opts);
            setFormValue(`${prefix}.pathReplaceIsRegex`, false, opts);
            setFormValue(`${prefix}.pathReplaceWith`, "", opts);
        } else if (nextMode === "stripPrefix") {
            setFormValue(`${prefix}.prefixAdd`, "", opts);
            setFormValue(`${prefix}.pathReplace`, "", opts);
            setFormValue(`${prefix}.pathReplaceIsRegex`, false, opts);
            setFormValue(`${prefix}.pathReplaceWith`, "", opts);
        } else {
            setFormValue(`${prefix}.prefixAdd`, "", opts);
            setFormValue(`${prefix}.prefixStrip`, "", opts);
            setFormValue(`${prefix}.prefixStripIsRegex`, false, opts);
        }
    }

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
                        Path Rewrite Configuration (docs
                        <span className="ml-1">
                            <a
                                className="text-xs text-blue-500 hover:text-blue-600"
                                href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/addprefix/"
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={e => {
                                    e.stopPropagation();
                                }}
                            >
                                [1]
                            </a>{" "}
                            ,
                            <a
                                className="text-xs text-blue-500 hover:text-blue-600 ml-1"
                                href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/stripprefix/"
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={e => {
                                    e.stopPropagation();
                                }}
                            >
                                [2]
                            </a>{" "}
                            ,
                            <a
                                className="text-xs text-blue-500 hover:text-blue-600 ml-1"
                                href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/replacepath/"
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={e => {
                                    e.stopPropagation();
                                }}
                            >
                                [3]
                            </a>{" "}
                            ,
                            <a
                                className="text-xs text-blue-500 hover:text-blue-600 ml-1"
                                href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/replacepathregex/"
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={e => {
                                    e.stopPropagation();
                                }}
                            >
                                [4]
                            </a>
                        </span>{" "}
                        )
                    </button>
                </CollapsibleTrigger>
                {onRemove && (
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
                )}
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
                        <>
                            <InfoBlock title="Path Rewrite Mode">
                                <Tabs
                                    value={mode}
                                    onValueChange={value => {
                                        handleModeChange(value as PathRewriteMode);
                                    }}
                                    className="w-fit"
                                >
                                    <TabsList>
                                        <TabsTrigger
                                            value="addPrefix"
                                            disabled={readOnly}
                                        >
                                            Add Prefix
                                        </TabsTrigger>
                                        <TabsTrigger
                                            value="stripPrefix"
                                            disabled={readOnly}
                                        >
                                            Strip Prefix
                                        </TabsTrigger>
                                        <TabsTrigger
                                            value="replace"
                                            disabled={readOnly}
                                        >
                                            Replace
                                        </TabsTrigger>
                                    </TabsList>
                                </Tabs>
                            </InfoBlock>

                            {mode === "addPrefix" && (
                                <InfoBlock title="Add Prefix">
                                    <Input
                                        value={prefixAdd.value}
                                        onChange={v => {
                                            prefixAdd.onChange(v);
                                        }}
                                        placeholder="/foo"
                                        className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                        disabled={readOnly}
                                    />
                                </InfoBlock>
                            )}

                            {mode === "stripPrefix" && (
                                <InfoBlock title="Strip Prefix">
                                    <div className="flex flex-wrap items-center gap-4">
                                        <Input
                                            value={prefixStrip.value}
                                            onChange={v => {
                                                prefixStrip.onChange(v);
                                            }}
                                            placeholder="/foo"
                                            className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                            disabled={readOnly}
                                        />
                                        <label
                                            htmlFor={`${prefix}-prefixStripIsRegex`}
                                            className="flex items-center gap-2 text-sm font-medium"
                                        >
                                            <Checkbox
                                                id={`${prefix}-prefixStripIsRegex`}
                                                checked={Boolean(prefixStripIsRegex.value)}
                                                onCheckedChange={checked => {
                                                    if (readOnly) {
                                                        return;
                                                    }

                                                    prefixStripIsRegex.onChange(checked === true);
                                                }}
                                                disabled={readOnly}
                                            />
                                            Is Regex
                                        </label>
                                    </div>
                                </InfoBlock>
                            )}

                            {mode === "replace" && (
                                <>
                                    <InfoBlock title="Replace Path">
                                        <div className="flex flex-wrap items-center gap-4">
                                            <Input
                                                value={pathReplace.value}
                                                onChange={v => {
                                                    pathReplace.onChange(v);
                                                }}
                                                placeholder="/foo"
                                                className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                                disabled={readOnly}
                                            />
                                            <label
                                                htmlFor={`${prefix}-pathReplaceIsRegex`}
                                                className="flex items-center gap-2 text-sm font-medium"
                                            >
                                                <Checkbox
                                                    id={`${prefix}-pathReplaceIsRegex`}
                                                    checked={Boolean(pathReplaceIsRegex.value)}
                                                    onCheckedChange={checked => {
                                                        if (readOnly) {
                                                            return;
                                                        }

                                                        pathReplaceIsRegex.onChange(checked === true);
                                                    }}
                                                    disabled={readOnly}
                                                />
                                                Is Regex
                                            </label>
                                        </div>
                                    </InfoBlock>
                                    <InfoBlock title="Replace Path With">
                                        <Input
                                            value={pathReplaceWith.value}
                                            onChange={v => {
                                                pathReplaceWith.onChange(v);
                                            }}
                                            placeholder="/bar"
                                            className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                            disabled={readOnly}
                                        />
                                    </InfoBlock>
                                </>
                            )}
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
