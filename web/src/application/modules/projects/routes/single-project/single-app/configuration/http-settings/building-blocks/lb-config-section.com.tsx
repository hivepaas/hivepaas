import { useEffect, useState } from "react";

import { Button, FieldError } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";
import { ELBStrategy } from "~/projects/module-shared/enums";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

const DEFAULT_STRATEGY_VALUE = "__default__";

interface LBConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    onRemove?: () => void;
}

export function LBConfigSection({ prefix, autoExpandToken, onRemove }: LBConfigSectionProps) {
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

    const {
        field: strategy,
        fieldState: { error: strategyError, invalid: isStrategyInvalid },
    } = useController({ control, name: `${prefix}.strategy` as never });

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
                        Load Balancing Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/load-balancing/service/"
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
                    <InfoBlock title="Strategy">
                        <Select
                            value={strategy.value === "" ? DEFAULT_STRATEGY_VALUE : strategy.value}
                            onValueChange={value => {
                                strategy.onChange(value === DEFAULT_STRATEGY_VALUE ? "" : value);
                            }}
                        >
                            <SelectTrigger
                                className="max-w-[320px]"
                                aria-invalid={isStrategyInvalid}
                            >
                                <SelectValue placeholder="Select strategy" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value={DEFAULT_STRATEGY_VALUE}>Default</SelectItem>
                                <SelectItem value={ELBStrategy.WeightedRoundRobin}>
                                    Weighted Round Robin (wrr)
                                </SelectItem>
                                <SelectItem value={ELBStrategy.PowerOfTwoChoices}>
                                    Power of Two Choices (p2c)
                                </SelectItem>
                                <SelectItem value={ELBStrategy.HighestRandomWeight}>
                                    Highest Random Weight (hrw)
                                </SelectItem>
                                <SelectItem value={ELBStrategy.LeastTime}>Least-Time</SelectItem>
                            </SelectContent>
                        </Select>
                        <FieldError errors={[strategyError]} />
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
