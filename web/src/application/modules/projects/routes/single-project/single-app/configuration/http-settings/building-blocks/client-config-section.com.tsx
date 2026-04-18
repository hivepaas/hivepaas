import { useState } from "react";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, Plus, Trash2, X } from "lucide-react";
import { useController, useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputNumberWithAddon } from "@application/shared/components";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
} from "../schemas";

interface ClientConfigSectionProps {
    prefix: string;
    onRemove?: () => void;
}

export function ClientConfigSection({ prefix, onRemove }: ClientConfigSectionProps) {
    const [open, setOpen] = useState(false);
    const [ipInput, setIpInput] = useState("");

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const enabledName = `${prefix}.enabled` as `domains.${number}.clientConfig.enabled`;
    const maxRequestBodyName = `${prefix}.maxRequestBody` as `domains.${number}.clientConfig.maxRequestBody`;
    const memRequestBodyName = `${prefix}.memRequestBody` as `domains.${number}.clientConfig.memRequestBody`;
    const allowedIPsName = `${prefix}.allowedIPs` as `domains.${number}.clientConfig.allowedIPs`;

    const { field: enabled } = useController({ control, name: enabledName as never });
    const {
        field: maxRequestBody,
        fieldState: { error: maxRequestBodyError },
    } = useController({ control, name: maxRequestBodyName as never });
    const {
        field: memRequestBody,
        fieldState: { error: memRequestBodyError },
    } = useController({ control, name: memRequestBodyName as never });

    const { fields: ipFields, append: appendIp, remove: removeIp } = useFieldArray({
        control,
        name: allowedIPsName as never,
    });

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
                        Client Configuration
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

                    <InfoBlock title="Max Request Body Size">
                        <InputNumberWithAddon
                            addonLeft="Bytes"
                            value={maxRequestBody.value ?? 0}
                            onValueChange={v => {
                                maxRequestBody.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[300px]"
                        />
                        <FieldError errors={[maxRequestBodyError]} />
                    </InfoBlock>

                    <InfoBlock title="Mem Request Body Size">
                        <InputNumberWithAddon
                            addonLeft="Bytes"
                            value={memRequestBody.value ?? 0}
                            onValueChange={v => {
                                memRequestBody.onChange(v ?? 0);
                            }}
                            useGrouping={false}
                            classNameContainer="max-w-[300px]"
                        />
                        <FieldError errors={[memRequestBodyError]} />
                    </InfoBlock>

                    <InfoBlock title="Allowed IPs">
                        <div className="flex flex-col gap-2 max-w-[400px]">
                            <div className="flex gap-2">
                                <Input
                                    value={ipInput}
                                    onChange={e => {
                                        setIpInput(e.target.value);
                                    }}
                                    placeholder="192.168.1.0/24"
                                    className="flex-1"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        if (ipInput.trim()) {
                                            appendIp({ value: ipInput.trim() });
                                            setIpInput("");
                                        }
                                    }}
                                >
                                    <Plus className="size-4" />
                                </Button>
                            </div>
                            <div className="flex flex-col divide-y">
                                {ipFields.map((field, index) => (
                                    <div
                                        key={field.id}
                                        className="flex items-center justify-between py-1.5"
                                    >
                                        <span className="font-mono text-sm">
                                            {(field as unknown as { value: string }).value}
                                        </span>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => {
                                                removeIp(index);
                                            }}
                                        >
                                            <Trash2 className="size-3.5" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
