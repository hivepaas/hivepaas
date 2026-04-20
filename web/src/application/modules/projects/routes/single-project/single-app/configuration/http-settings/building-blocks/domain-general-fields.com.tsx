import { Checkbox, FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { ContainerPort, SslCert } from "../form-components";
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

    const {
        field: domainRedirect,
        fieldState: { error: domainRedirectError },
    } = useController({ control, name: `${p}.domainRedirect` });
    const { field: forceHttps } = useController({ control, name: `${p}.forceHttps` });

    return (
        <>
            <ContainerPort domainIndex={domainIndex} />
            <SslCert domainIndex={domainIndex} />

            <InfoBlock title="Redirect To">
                <Input
                    {...domainRedirect}
                    value={domainRedirect.value}
                    onChange={domainRedirect.onChange}
                    placeholder="https://other-domain.com"
                    className="max-w-[400px]"
                />
                <FieldError errors={[domainRedirectError]} />
            </InfoBlock>

            <InfoBlock title="Force HTTPS">
                <Checkbox
                    checked={forceHttps.value}
                    onCheckedChange={forceHttps.onChange}
                />
            </InfoBlock>
        </>
    );
}
