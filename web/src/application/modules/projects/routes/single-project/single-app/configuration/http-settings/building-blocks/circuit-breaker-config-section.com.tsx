import { useEffect, useState } from "react";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { InputNumber } from "@components/ui/input-number";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface CircuitBreakerConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    readOnly?: boolean;
    onRemove?: () => void;
}

export function CircuitBreakerConfigSection({
    prefix,
    autoExpandToken,
    readOnly = false,
    onRemove,
}: CircuitBreakerConfigSectionProps) {
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
        field: expression,
        fieldState: { error: expressionError, invalid: expressionInvalid },
    } = useController({ control, name: `${prefix}.expression` as never });
    const {
        field: checkPeriod,
        fieldState: { error: checkPeriodError, invalid: checkPeriodInvalid },
    } = useController({ control, name: `${prefix}.checkPeriod` as never });
    const {
        field: fallbackDuration,
        fieldState: { error: fallbackDurationError, invalid: fallbackDurationInvalid },
    } = useController({ control, name: `${prefix}.fallbackDuration` as never });
    const {
        field: recoveryDuration,
        fieldState: { error: recoveryDurationError, invalid: recoveryDurationInvalid },
    } = useController({ control, name: `${prefix}.recoveryDuration` as never });
    const {
        field: responseCode,
        fieldState: { error: responseCodeError, invalid: responseCodeInvalid },
    } = useController({ control, name: `${prefix}.responseCode` as never });
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
                        Circuit Breaker Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/circuitbreaker/"
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
                            <InfoBlock title="Expression">
                                <Input
                                    value={expression.value}
                                    onChange={v => {
                                        expression.onChange(v);
                                    }}
                                    placeholder="NetworkErrorRatio() > 0.30"
                                    className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                    aria-invalid={expressionInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[expressionError]} />
                            </InfoBlock>

                            <InfoBlock title="Check Period">
                                <Input
                                    value={checkPeriod.value}
                                    onChange={v => {
                                        checkPeriod.onChange(v);
                                    }}
                                    placeholder="100ms"
                                    className="max-w-[100px]"
                                    aria-invalid={checkPeriodInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[checkPeriodError]} />
                            </InfoBlock>

                            <InfoBlock title="Fallback Duration">
                                <Input
                                    value={fallbackDuration.value}
                                    onChange={v => {
                                        fallbackDuration.onChange(v);
                                    }}
                                    placeholder="10s"
                                    className="max-w-[100px]"
                                    aria-invalid={fallbackDurationInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[fallbackDurationError]} />
                            </InfoBlock>

                            <InfoBlock title="Recovery Duration">
                                <Input
                                    value={recoveryDuration.value}
                                    onChange={v => {
                                        recoveryDuration.onChange(v);
                                    }}
                                    placeholder="10s"
                                    className="max-w-[100px]"
                                    aria-invalid={recoveryDurationInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[recoveryDurationError]} />
                            </InfoBlock>

                            <InfoBlock title="Response Code">
                                <InputNumber
                                    value={responseCode.value}
                                    onValueChange={responseCode.onChange}
                                    placeholder="503"
                                    className="max-w-[100px]"
                                    useGrouping={false}
                                    aria-invalid={responseCodeInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[responseCodeError]} />
                            </InfoBlock>
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
