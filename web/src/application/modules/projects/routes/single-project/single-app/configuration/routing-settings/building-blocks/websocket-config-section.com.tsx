import { useEffect, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface WebsocketConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    readOnly?: boolean;
    onRemove?: () => void;
}

export function WebsocketConfigSection({
    prefix,
    autoExpandToken,
    readOnly = false,
    onRemove,
}: WebsocketConfigSectionProps) {
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
            <div className="sticky top-11 z-10 flex justify-between items-center font-medium bg-accent py-2 px-3 rounded-lg shadow-xs">
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
                        Websocket Configuration
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
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
