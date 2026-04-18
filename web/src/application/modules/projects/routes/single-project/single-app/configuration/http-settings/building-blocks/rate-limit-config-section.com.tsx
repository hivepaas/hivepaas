import { useState } from "react";

import { Button, Checkbox, FieldError } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, InputNumberWithAddon } from "@application/shared/components";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
} from "../schemas";

interface RateLimitConfigSectionProps {
    prefix: string;
    onRemove?: () => void;
}

export function RateLimitConfigSection({ prefix, onRemove }: RateLimitConfigSectionProps) {
    const [open, setOpen] = useState(false);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const { field: enabled } = useController({ control, name: `${prefix}.enabled` as never });
    const {
        field: average,
        fieldState: { error: averageError },
    } = useController({ control, name: `${prefix}.average` as never });
    const {
        field: period,
        fieldState: { error: periodError },
    } = useController({ control, name: `${prefix}.period` as never });
    const {
        field: burst,
        fieldState: { error: burstError },
    } = useController({ control, name: `${prefix}.burst` as never });
    const {
        field: maxInFlightReq,
        fieldState: { error: maxInFlightReqError },
    } = useController({ control, name: `${prefix}.maxInFlightReq` as never });

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
                        {open ? <ChevronDown className="size-4 shrink-0" /> : <ChevronRight className="size-4 shrink-0" />}
                        Rate Limit Configuration
                    </button>
                </CollapsibleTrigger>
                {onRemove && (
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
                )}
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Enabled">
                        <Checkbox
                            checked={enabled.value ?? false}
                            onCheckedChange={enabled.onChange}
                        />
                    </InfoBlock>

                    <InfoBlock title="Average">
                        <InputNumberWithAddon
                            addonLeft="req/s"
                            value={average.value ?? 0}
                            onValueChange={v => {
                                average.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[240px]"
                        />
                        <FieldError errors={[averageError]} />
                    </InfoBlock>

                    <InfoBlock title="Period">
                        <InputNumberWithAddon
                            addonLeft="ms"
                            value={period.value ?? 0}
                            onValueChange={v => {
                                period.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[240px]"
                        />
                        <FieldError errors={[periodError]} />
                    </InfoBlock>

                    <InfoBlock title="Burst">
                        <InputNumberWithAddon
                            addonLeft="req"
                            value={burst.value ?? 0}
                            onValueChange={v => {
                                burst.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[240px]"
                        />
                        <FieldError errors={[burstError]} />
                    </InfoBlock>

                    <InfoBlock title="Max In-Flight Requests">
                        <InputNumberWithAddon
                            addonLeft="req"
                            value={maxInFlightReq.value ?? 0}
                            onValueChange={v => {
                                maxInFlightReq.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[240px]"
                        />
                        <FieldError errors={[maxInFlightReqError]} />
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
