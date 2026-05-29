import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

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

    const { field: forceHttps } = useController({ control, name: `${p}.forceHttps` });

    return (
        <ContentBlock label="General">
            <div className="flex flex-col gap-6">
                <ContainerPort
                    domainIndex={domainIndex}
                    readOnly={readOnly}
                />
                <SslCert
                    domainIndex={domainIndex}
                    readOnly={readOnly}
                />

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
                <RedirectTo
                    domainIndex={domainIndex}
                    readOnly={readOnly}
                />
            </div>
        </ContentBlock>
    );
}
