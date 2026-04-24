import { useState } from "react";

import { Button, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Textarea } from "@components/ui/textarea";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface ClientConfigSectionProps {
    prefix: string;
    onRemove?: () => void;
}

export function ClientConfigSection({ prefix, onRemove }: ClientConfigSectionProps) {
    const [open, setOpen] = useState(false);

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
                    <InfoBlock title="Max Request Body Size">
                        <Input
                            value={maxRequestBody.value}
                            onChange={maxRequestBody.onChange}
                            className="max-w-[100px]"
                            aria-invalid={isMaxRequestBodyInvalid}
                        />
                        <FieldError errors={[maxRequestBodyError]} />
                    </InfoBlock>

                    <InfoBlock title="Mem Request Body Size">
                        <Input
                            value={memRequestBody.value}
                            onChange={memRequestBody.onChange}
                            className="max-w-[100px]"
                            aria-invalid={isMemRequestBodyInvalid}
                        />
                        <FieldError errors={[memRequestBodyError]} />
                    </InfoBlock>

                    <InfoBlock title="Allowed IPs">
                        <div className="flex flex-col gap-2 max-w-[400px]">
                            <Textarea
                                {...allowedIPs}
                                onChange={allowedIPs.onChange}
                                placeholder="192.168.1.0/24"
                                rows={2}
                                className="resize-y"
                                aria-invalid={isAllowedIPsInvalid}
                            />
                            <FieldError errors={[allowedIPsError]} />
                        </div>
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
