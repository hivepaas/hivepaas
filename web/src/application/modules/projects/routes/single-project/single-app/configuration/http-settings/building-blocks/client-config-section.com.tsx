import { useEffect, useState } from "react";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Textarea } from "@components/ui/textarea";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface ClientConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    readOnly?: boolean;
    onRemove?: () => void;
}

export function ClientConfigSection({ prefix, autoExpandToken, readOnly = false, onRemove }: ClientConfigSectionProps) {
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

    const maxRequestBodyName = `${prefix}.maxRequestBody` as `domains.${number}.clientConfig.maxRequestBody`;
    const memRequestBodyName = `${prefix}.memRequestBody` as `domains.${number}.clientConfig.memRequestBody`;
    const allowedIPsName = `${prefix}.allowedIPs` as `domains.${number}.clientConfig.allowedIPs`;

    const {
        field: maxRequestBody,
        fieldState: { error: maxRequestBodyError, invalid: isMaxRequestBodyInvalid },
    } = useController({ control, name: maxRequestBodyName as never });
    const {
        field: memRequestBody,
        fieldState: { error: memRequestBodyError, invalid: isMemRequestBodyInvalid },
    } = useController({ control, name: memRequestBodyName as never });

    const {
        field: allowedIPs,
        fieldState: { error: allowedIPsError, invalid: isAllowedIPsInvalid },
    } = useController({ control, name: allowedIPsName as never });
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
                        Client Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/buffering/"
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
                            <InfoBlock title="Max Request Body Size">
                                <Input
                                    value={maxRequestBody.value}
                                    onChange={maxRequestBody.onChange}
                                    className="max-w-[100px]"
                                    aria-invalid={isMaxRequestBodyInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[maxRequestBodyError]} />
                            </InfoBlock>

                            <InfoBlock title="Mem Request Body Size">
                                <Input
                                    value={memRequestBody.value}
                                    onChange={memRequestBody.onChange}
                                    className="max-w-[100px]"
                                    aria-invalid={isMemRequestBodyInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[memRequestBodyError]} />
                            </InfoBlock>

                            <InfoBlock title="Allowed IPs">
                                <div className={`flex flex-col gap-2 ${HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}`}>
                                    <Textarea
                                        {...allowedIPs}
                                        onChange={allowedIPs.onChange}
                                        placeholder="192.168.1.0/24"
                                        rows={2}
                                        className="resize-y"
                                        aria-invalid={isAllowedIPsInvalid}
                                        disabled={readOnly}
                                    />
                                    <FieldError errors={[allowedIPsError]} />
                                </div>
                            </InfoBlock>
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
