import { Checkbox, Tabs, TabsList, TabsTrigger } from "@components/ui";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useController, useFormContext } from "react-hook-form";
import { ERoutingProtocol } from "~/projects/module-shared/enums";

import { ContentBlock, InfoBlock } from "@application/shared/components";

import { ContainerPort, RedirectTo, SslCert } from "../form-components";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

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
    const { field: domain } = useController({ control, name: `${p}.domain` });

    const currentProtocol = protocol.value;
    const isHttp = currentProtocol === ERoutingProtocol.HTTP;
    const isUdp = currentProtocol === ERoutingProtocol.UDP;
    const domainName = domain.value.trim();

    return (
        <ContentBlock label={<span className="text-red-500">Selected domain: {domainName || "—"}</span>}>
            <div className="flex flex-col gap-6">
                <InfoBlock title="Protocol">
                    <Tabs
                        value={currentProtocol}
                        onValueChange={value => {
                            if (readOnly) {
                                return;
                            }
                            protocol.onChange(value);
                        }}
                    >
                        <TabsList>
                            <TabsTrigger
                                value={ERoutingProtocol.HTTP}
                                disabled={readOnly}
                            >
                                HTTP
                            </TabsTrigger>
                            <TabsTrigger
                                value={ERoutingProtocol.TCP}
                                disabled={readOnly}
                            >
                                TCP
                            </TabsTrigger>
                            <TabsTrigger
                                value={ERoutingProtocol.UDP}
                                disabled={readOnly}
                            >
                                UDP
                            </TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {isUdp && (
                    <div className={cn(dashedBorderBox, "text-sm leading-6")}>
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

                        <InfoBlock title="TLS Passthrough">
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
                            <InfoBlock title="Force HTTPS">
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
        </ContentBlock>
    );
}
