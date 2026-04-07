import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";
import { EEndpointResolutionMode, EPortConfigProtocol, EPortConfigPublishMode } from "~/projects/module-shared/enums";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

export function EndpointPortConfigFields() {
    const { control, watch, setValue } = useFormContext<
        AppConfigNetworksFormSchemaInput,
        unknown,
        AppConfigNetworksFormSchemaOutput
    >();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "portConfigs",
    });

    const [published, setPublished] = useState("");
    const [target, setTarget] = useState("");
    const [protocol, setProtocol] = useState<EPortConfigProtocol>(EPortConfigProtocol.TCP);
    const [publishMode, setPublishMode] = useState<EPortConfigPublishMode>(EPortConfigPublishMode.Host);
    const resolutionMode = watch("resolutionMode");

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Resolution Mode"
                        content="Service discovery mode for this app."
                    />
                }
            >
                <div className="flex flex-col gap-4">
                    <div className="flex flex-col gap-2">
                        <Tabs
                            value={resolutionMode}
                            onValueChange={value => {
                                setValue("resolutionMode", value as EEndpointResolutionMode);
                            }}
                            className="w-fit"
                        >
                            <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                                <TabsTrigger value={EEndpointResolutionMode.VIP}>VIP</TabsTrigger>
                                <TabsTrigger value={EEndpointResolutionMode.DNSRR}>DNSRR</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </div>
                </div>
            </InfoBlock>

            <div className={cn(dashedBorderBox, "text-center")}>
                <div className="max-w-[735px] mx-auto">
                    <span className="text-sm text-orange-500">Note:</span> It is recommended not to expose container
                    ports directly to the host to avoid conflict between apps from different projects. If you need
                    external access, consider using port mapping in the HTTP settings.
                </div>
            </div>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Port Config"
                        content="Published host ports mapped to container target ports."
                    />
                }
            >
                <div className="flex flex-col gap-2">
                    <div className="flex gap-3 items-center max-w-[800px]">
                        <InputWithAddOn
                            addonLeft="Host"
                            value={published}
                            onChange={e => {
                                setPublished(e.target.value);
                            }}
                            placeholder="8000"
                            classNameContainer="flex-1"
                        />
                        <InputWithAddOn
                            addonLeft="Container"
                            value={target}
                            onChange={e => {
                                setTarget(e.target.value);
                            }}
                            placeholder="80"
                            classNameContainer="flex-1"
                        />
                        <div className="flex items-center rounded-md border border-input h-9 flex-1">
                            <span className="px-3 text-sm border-r border-input bg-muted/50 h-full flex items-center">
                                Protocol
                            </span>
                            <Select
                                value={protocol}
                                onValueChange={value => {
                                    setProtocol(value as EPortConfigProtocol);
                                }}
                            >
                                <SelectTrigger className="w-[80px] border-0 shadow-none rounded-l-none focus:ring-0">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={EPortConfigProtocol.TCP}>TCP</SelectItem>
                                    <SelectItem value={EPortConfigProtocol.UDP}>UDP</SelectItem>
                                    <SelectItem value={EPortConfigProtocol.SCTP}>SCTP</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="flex items-center rounded-md border border-input h-9 flex-1">
                            <span className="px-3 text-sm border-r border-input bg-muted/50 h-full flex items-center">
                                Mode
                            </span>
                            <Select
                                value={publishMode}
                                onValueChange={value => {
                                    setPublishMode(value as EPortConfigPublishMode);
                                }}
                            >
                                <SelectTrigger className="w-[80px] border-0 shadow-none rounded-l-none focus:ring-0">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={EPortConfigPublishMode.Host}>Host</SelectItem>
                                    <SelectItem value={EPortConfigPublishMode.Ingress}>Ingress</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => {
                                append({
                                    published: published.trim(),
                                    target: target.trim(),
                                    protocol,
                                    publishMode,
                                });
                                setPublished("");
                                setTarget("");
                            }}
                            disabled={!published.trim() || !target.trim()}
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>
                </div>

                <div className="divide-y divide-zinc-200">
                    {fields.map((field, index) => (
                        <div
                            key={field.id}
                            className="flex items-center gap-3 py-2 max-w-[800px]"
                        >
                            <Input
                                value={field.published}
                                disabled
                            />
                            <Input
                                value={field.target}
                                disabled
                            />
                            <Input
                                value={field.protocol}
                                disabled
                            />
                            <Input
                                value={field.publishMode}
                                disabled
                            />
                            <div className="w-[76px]">
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => {
                                        remove(index);
                                    }}
                                >
                                    <Trash2 className="size-4" />
                                </Button>
                            </div>
                        </div>
                    ))}
                </div>
            </InfoBlock>
        </>
    );
}
