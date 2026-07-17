import { Checkbox, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent } from "@components/ui/collapsible";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import type { HivePaaSHttpSettingsFormInput, HivePaaSHttpSettingsFormOutput } from "../schemas";

interface RateLimitConfigSectionProps {
    prefix: `domains.${number}.rateLimitConfig`;
    readOnly?: boolean;
}

export function RateLimitConfigSection({ prefix, readOnly = false }: RateLimitConfigSectionProps) {
    const { control } = useFormContext<HivePaaSHttpSettingsFormInput, unknown, HivePaaSHttpSettingsFormOutput>();

    const {
        field: average,
        fieldState: { error: averageError, invalid: averageInvalid },
    } = useController({ control, name: `${prefix}.average` });
    const {
        field: period,
        fieldState: { error: periodError, invalid: periodInvalid },
    } = useController({ control, name: `${prefix}.period` });
    const {
        field: burst,
        fieldState: { error: burstError, invalid: burstInvalid },
    } = useController({ control, name: `${prefix}.burst` });
    const {
        field: maxInFlightReq,
        fieldState: { error: maxInFlightReqError, invalid: maxInFlightReqInvalid },
    } = useController({ control, name: `${prefix}.maxInFlightReq` });
    const { field: enabled } = useController({ control, name: `${prefix}.enabled` });
    const isEnabled = enabled.value;

    return (
        <Collapsible open>
            <div className="flex justify-between items-center font-medium bg-accent py-2 px-3 rounded-lg">
                <div className="flex min-w-0 flex-1 items-center gap-2 text-sm font-medium">
                    <span>Rate Limit Configuration (docs</span>
                    <a
                        className="text-xs text-blue-500 hover:text-blue-600"
                        href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/ratelimit/"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        [1]
                    </a>
                    <span>,</span>
                    <a
                        className="text-xs text-blue-500 hover:text-blue-600"
                        href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/inflightreq/"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        [2]
                    </a>
                    <span>)</span>
                </div>
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

                                enabled.onChange(value === true);
                            }}
                            disabled={readOnly}
                        />
                    </InfoBlock>
                    {isEnabled && (
                        <>
                            <InfoBlock title="Average">
                                <InputNumber
                                    value={average.value}
                                    onValueChange={average.onChange}
                                    placeholder="10"
                                    className="max-w-[100px]"
                                    useGrouping={false}
                                    aria-invalid={averageInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[averageError]} />
                            </InfoBlock>

                            <InfoBlock title="Burst">
                                <InputNumber
                                    value={burst.value}
                                    onValueChange={burst.onChange}
                                    placeholder="20"
                                    className="max-w-[100px]"
                                    useGrouping={false}
                                    aria-invalid={burstInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[burstError]} />
                            </InfoBlock>

                            <InfoBlock title="Period">
                                <Input
                                    value={period.value}
                                    onChange={period.onChange}
                                    placeholder="1m"
                                    className="max-w-[100px]"
                                    aria-invalid={periodInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[periodError]} />
                            </InfoBlock>

                            <InfoBlock title="Max InFlight Requests">
                                <InputNumber
                                    value={maxInFlightReq.value}
                                    onValueChange={maxInFlightReq.onChange}
                                    placeholder="10"
                                    className="max-w-[100px]"
                                    useGrouping={false}
                                    aria-invalid={maxInFlightReqInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[maxInFlightReqError]} />
                            </InfoBlock>
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
