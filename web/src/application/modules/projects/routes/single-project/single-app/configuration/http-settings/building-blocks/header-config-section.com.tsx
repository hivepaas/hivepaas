import { useEffect, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList, SingleValueList } from "@application/shared/form";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface HeaderConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    onRemove?: () => void;
}

export function HeaderConfigSection({ prefix, autoExpandToken, onRemove }: HeaderConfigSectionProps) {
    const [open, setOpen] = useState(false);
    useEffect(() => {
        if (autoExpandToken !== undefined) {
            setOpen(true);
        }
    }, [autoExpandToken]);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const { field: enabled } = useController({ control, name: `${prefix}.enabled` as never });
    const isEnabled = Boolean(enabled.value);

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
                        Header Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/headers/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            (docs)
                        </a>
                    </button>
                </CollapsibleTrigger>
                {onRemove && (
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="shrink-0 text-muted-foreground hover:text-destructive h-fit"
                        title="Remove section"
                        onClick={onRemove}
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
                            onCheckedChange={enabled.onChange}
                        />
                    </InfoBlock>
                    {isEnabled && (
                        <>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="To Add To Requests"
                                        content="Headers to add to the request."
                                    />
                                }
                            >
                                <KeyValueList<AppConfigHttpSettingsFormSchemaInput>
                                    name={`${prefix}.toAddToRequests` as never}
                                    className="max-w-[550px]"
                                />
                            </InfoBlock>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="To Remove From Requests"
                                        content="Headers to remove from the request."
                                    />
                                }
                            >
                                <SingleValueList<AppConfigHttpSettingsFormSchemaInput>
                                    name={`${prefix}.toRemoveFromRequests` as never}
                                    placeholder="Header name"
                                    className="max-w-[314px]"
                                />
                            </InfoBlock>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="To Add To Responses"
                                        content="Headers to add to the response."
                                    />
                                }
                            >
                                <KeyValueList<AppConfigHttpSettingsFormSchemaInput>
                                    name={`${prefix}.toAddToResponses` as never}
                                    className="max-w-[550px]"
                                />
                            </InfoBlock>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="To Remove From Responses"
                                        content="Headers to remove from the response."
                                    />
                                }
                            >
                                <SingleValueList<AppConfigHttpSettingsFormSchemaInput>
                                    name={`${prefix}.toRemoveFromResponses` as never}
                                    placeholder="Header name"
                                    className="max-w-[314px]"
                                />
                            </InfoBlock>
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
