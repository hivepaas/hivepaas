import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { ContainerPort, RedirectTo, SslCert } from "../form-components";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface DomainGeneralFieldsProps {
    domainIndex: number;
}

export function DomainGeneralFields({ domainIndex }: DomainGeneralFieldsProps) {
    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const p = `domains.${domainIndex}` as const;

    const { field: forceHttps } = useController({ control, name: `${p}.forceHttps` });

    return (
        <>
            <ContainerPort domainIndex={domainIndex} />
            <SslCert domainIndex={domainIndex} />

            <InfoBlock title="Force HTTPS">
                <Checkbox
                    checked={forceHttps.value}
                    onCheckedChange={forceHttps.onChange}
                />
            </InfoBlock>
            <RedirectTo domainIndex={domainIndex} />
        </>
    );
}
