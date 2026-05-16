import { FieldError } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { useController, useFormContext } from "react-hook-form";
import { ELBStrategy } from "~/projects/module-shared/enums";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

const DEFAULT_STRATEGY_VALUE = "__default__";

interface LBConfigSectionProps {
    prefix: string;
}

export function LBConfigSection({ prefix }: LBConfigSectionProps) {
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
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">
                Load Balancing Configuration{" "}
                <a
                    className="text-xs text-blue-500 hover:text-blue-600"
                    href="https://doc.traefik.io/traefik/reference/routing-configuration/http/load-balancing/service/"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    (docs)
                </a>
            </h3>
            <div className="flex flex-col gap-6 px-2">
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
                            <SelectItem value={ELBStrategy.WeightedRoundRobin}>Weighted Round Robin (wrr)</SelectItem>
                            <SelectItem value={ELBStrategy.PowerOfTwoChoices}>Power of Two Choices (p2c)</SelectItem>
                            <SelectItem value={ELBStrategy.HighestRandomWeight}>Highest Random Weight (hrw)</SelectItem>
                            <SelectItem value={ELBStrategy.LeastTime}>Least-Time</SelectItem>
                        </SelectContent>
                    </Select>
                    <FieldError errors={[strategyError]} />
                </InfoBlock>
            </div>
        </>
        // </Collapsible>
    );
}
