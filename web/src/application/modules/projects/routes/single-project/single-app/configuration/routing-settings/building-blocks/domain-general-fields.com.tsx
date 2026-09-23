import { Checkbox } from "@components/ui";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Cable, Globe, Radio } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components";
import { ERoutingProtocol } from "~/projects/module-shared/enums";

import { InfoBlock } from "@application/shared/components";

import { ContainerPort, RedirectTo, SslCert } from "../form-components";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

const PROTOCOL_OPTIONS: OptionCard<ERoutingProtocol>[] = [
    {
        value: ERoutingProtocol.HTTP,
        label: "HTTP",
        description: "Web traffic with domains, TLS and routing rules",
        icon: Globe,
    },
    {
        value: ERoutingProtocol.TCP,
        label: "TCP",
        description: "Raw TCP stream for databases, caches, and custom protocols",
        icon: Cable,
    },
    {
        value: ERoutingProtocol.UDP,
        label: "UDP",
        description: "Not routed by the proxy; publish a port in Networks",
        icon: Radio,
    },
];

interface DomainGeneralFieldsProps {
    domainIndex: number;
    readOnly?: boolean;
}

export function DomainGeneralFields({ domainIndex, readOnly = false }: DomainGeneralFieldsProps) {
    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const p = `domains.${domainIndex}` as const;

    const { field: protocol } = useController({ control, name: `${p}.protocol` });
    const { field: tlsPassthrough } = useController({ control, name: `${p}.tlsPassthrough` });
    const { field: forceHttps } = useController({ control, name: `${p}.forceHttps` });

    const currentProtocol = protocol.value;
    const isHttp = currentProtocol === ERoutingProtocol.HTTP;
    const isUdp = currentProtocol === ERoutingProtocol.UDP;

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={240}
                title="Protocol"
            >
                <OptionCardGroup
                    options={PROTOCOL_OPTIONS}
                    value={currentProtocol}
                    onChange={protocol.onChange}
                    readOnly={readOnly}
                    className="grid-cols-1 sm:grid-cols-3 max-w-[660px]"
                />
            </InfoBlock>

            {isUdp && (
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-orange-500">Note:</span> UDP routing is not supported by the
                    reverse proxy. To expose and route UDP traffic for this application, please configure directly
                    published container ports in the <strong>Networks</strong> tab instead.
                </div>
            )}

            {!isUdp && (
                <>
                    <ContainerPort
                        domainIndex={domainIndex}
                        readOnly={readOnly}
                    />

                    <SslCert
                        domainIndex={domainIndex}
                        readOnly={readOnly}
                    />

                    <InfoBlock
                        titleWidth={240}
                        title="TLS Passthrough"
                    >
                        <Checkbox
                            checked={tlsPassthrough.value}
                            onCheckedChange={value => {
                                if (readOnly) {
                                    return;
                                }
                                tlsPassthrough.onChange(value);
                            }}
                            disabled={readOnly}
                        />
                    </InfoBlock>

                    {isHttp && (
                        <InfoBlock
                            titleWidth={240}
                            title="Force HTTPS"
                        >
                            <Checkbox
                                checked={forceHttps.value}
                                onCheckedChange={value => {
                                    if (readOnly) {
                                        return;
                                    }

                                    forceHttps.onChange(value);
                                }}
                                disabled={readOnly}
                            />
                        </InfoBlock>
                    )}

                    {isHttp && !tlsPassthrough.value && (
                        <RedirectTo
                            domainIndex={domainIndex}
                            readOnly={readOnly}
                        />
                    )}
                </>
            )}
        </div>
    );
}
