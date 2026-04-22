import { useState } from "react";

import { Button, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { InputNumber } from "@components/ui/input-number";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

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

    const {
        field: average,
        fieldState: { error: averageError, invalid: averageInvalid },
    } = useController({ control, name: `${prefix}.average` as never });
    const {
        field: period,
        fieldState: { error: periodError, invalid: periodInvalid },
    } = useController({ control, name: `${prefix}.period` as never });
    const {
        field: burst,
        fieldState: { error: burstError, invalid: burstInvalid },
    } = useController({ control, name: `${prefix}.burst` as never });
    const {
        field: maxInFlightReq,
        fieldState: { error: maxInFlightReqError, invalid: maxInFlightReqInvalid },
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
                        {open ? (
                            <ChevronDown className="size-4 shrink-0" />
                        ) : (
                            <ChevronRight className="size-4 shrink-0" />
                        )}
                        Rate Limit Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/ratelimit/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            (docs 1)
                        </a>
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/inflightreq/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            (docs 2)
                        </a>
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
                    <InfoBlock title="Average">
                        <InputNumber
                            value={average.value}
                            onValueChange={average.onChange}
                            placeholder="100"
                            className="max-w-[100px]"
                            useGrouping={false}
                            aria-invalid={averageInvalid}
                        />
                        <FieldError errors={[averageError]} />
                    </InfoBlock>

                    <InfoBlock title="Period">
                        <Input
                            value={period.value}
                            onChange={v => {
                                period.onChange(v);
                            }}
                            placeholder="1s"
                            className="max-w-[100px]"
                            aria-invalid={periodInvalid}
                        />
                        <FieldError errors={[periodError]} />
                    </InfoBlock>

                    <InfoBlock title="Burst">
                        <InputNumber
                            value={burst.value}
                            onValueChange={burst.onChange}
                            placeholder="100"
                            className="max-w-[100px]"
                            useGrouping={false}
                            aria-invalid={burstInvalid}
                        />
                        <FieldError errors={[burstError]} />
                    </InfoBlock>

                    <InfoBlock title="Max In-Flight Requests">
                        <InputNumber
                            value={maxInFlightReq.value}
                            onValueChange={maxInFlightReq.onChange}
                            placeholder="100"
                            className="max-w-[100px]"
                            useGrouping={false}
                            aria-invalid={maxInFlightReqInvalid}
                        />
                        <FieldError errors={[maxInFlightReqError]} />
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
