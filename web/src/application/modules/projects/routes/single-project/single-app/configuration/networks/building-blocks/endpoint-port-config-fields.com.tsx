import { useState } from "react";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useFieldArray, useFormContext } from "react-hook-form";
import { EEndpointResolutionMode, EPortConfigProtocol, EPortConfigPublishMode } from "~/projects/module-shared/enums";

import { InfoBlock, InputNumberWithAddon, LabelWithInfo } from "@application/shared/components";
import { FieldListLayout } from "@application/shared/form";

import { type AppConfigNetworksFormSchemaInput, type AppConfigNetworksFormSchemaOutput } from "../schemas";

const portFieldsGridClass = "grid flex-1 min-w-0 w-full grid-cols-4 gap-3 items-center";

export function EndpointPortConfigFields({ readOnly = false }: Props) {
    const { control, watch, setValue } = useFormContext<
        AppConfigNetworksFormSchemaInput,
        unknown,
        AppConfigNetworksFormSchemaOutput
    >();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "portConfigs",
    });

    const [published, setPublished] = useState<number>(0);
    const [target, setTarget] = useState<number>(0);
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
                                if (readOnly) {
                                    return;
                                }

                                setValue("resolutionMode", value as EEndpointResolutionMode);
                            }}
                            className="w-fit"
                        >
                            <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                                <TabsTrigger
                                    value={EEndpointResolutionMode.VIP}
                                    disabled={readOnly}
                                >
                                    VIP
                                </TabsTrigger>
                                <TabsTrigger
                                    value={EEndpointResolutionMode.DNSRR}
                                    disabled={readOnly}
                                >
                                    DNSRR
                                </TabsTrigger>
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
                <FieldListLayout
                    className="max-w-[800px]"
                    inputRow={
                        <div className={portFieldsGridClass}>
                            <InputNumberWithAddon
                                addonLeft="Host"
                                value={published}
                                onValueChange={v => {
                                    if (readOnly) {
                                        return;
                                    }

                                    setPublished(v ?? 0);
                                }}
                                useGrouping={false}
                                placeholder="8000"
                                classNameContainer="min-w-0"
                                disabled={readOnly}
                            />
                            <InputNumberWithAddon
                                addonLeft="Container"
                                value={target}
                                onValueChange={v => {
                                    if (readOnly) {
                                        return;
                                    }

                                    setTarget(v ?? 0);
                                }}
                                useGrouping={false}
                                placeholder="80"
                                classNameContainer="min-w-0"
                                disabled={readOnly}
                            />
                            <div className="flex min-w-0 items-center rounded-md border border-input h-9">
                                <span className="px-3 text-sm border-r border-input bg-muted/50 h-full flex items-center">
                                    Protocol
                                </span>
                                <Select
                                    value={protocol}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        setProtocol(value as EPortConfigProtocol);
                                    }}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger className="w-[80px] flex-1 border-0 shadow-none rounded-l-none focus:ring-0">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={EPortConfigProtocol.TCP}>TCP</SelectItem>
                                        <SelectItem value={EPortConfigProtocol.UDP}>UDP</SelectItem>
                                        <SelectItem value={EPortConfigProtocol.SCTP}>SCTP</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="flex min-w-0 items-center rounded-md border border-input h-9">
                                <span className="px-3 text-sm border-r border-input bg-muted/50 h-full flex items-center">
                                    Mode
                                </span>
                                <Select
                                    value={publishMode}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        setPublishMode(value as EPortConfigPublishMode);
                                    }}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger className="w-[80px] flex-1 border-0 shadow-none rounded-l-none focus:ring-0">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={EPortConfigPublishMode.Host}>Host</SelectItem>
                                        <SelectItem value={EPortConfigPublishMode.Ingress}>Ingress</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>
                    }
                    onAdd={() => {
                        if (readOnly) {
                            return;
                        }

                        append({ published, target, protocol, publishMode });
                        setPublished(0);
                        setTarget(0);
                    }}
                    addDisabled={readOnly || published === 0 || target === 0}
                    disabled={readOnly}
                    items={fields.map((field, index) => ({
                        id: field.id,
                        content: (
                            <div className={portFieldsGridClass}>
                                <span className="text-sm wrap-break-word min-w-0">{field.published}</span>
                                <span className="text-sm wrap-break-word min-w-0">{field.target}</span>
                                <span className="text-sm wrap-break-word min-w-0">{field.protocol}</span>
                                <span className="text-sm wrap-break-word min-w-0">{field.publishMode}</span>
                            </div>
                        ),
                        onRemove: () => {
                            remove(index);
                        },
                    }))}
                />
            </InfoBlock>
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
