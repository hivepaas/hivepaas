import { useMemo, useState } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectSslCertQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { ContainerPort, SslCert } from "../form-components";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface DomainGeneralFieldsProps {
    domainIndex: number;
}

export function DomainGeneralFields({ domainIndex }: DomainGeneralFieldsProps) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const p = `domains.${domainIndex}` as const;

    const [searchQuery, setSearchQuery] = useState("");

    const {
        data: { data: sslCerts } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectSslCertQueries.useFindManyPaginated({ projectID: projectId, search: searchQuery });

    const {
        field: domainRedirect,
        fieldState: { error: domainRedirectError },
    } = useController({ control, name: `${p}.domainRedirect` });
    const { field: forceHttps } = useController({ control, name: `${p}.forceHttps` });
    const {
        field: sslCert,
        fieldState: { error: sslCertError, invalid: isSslCertInvalid },
    } = useController({ control, name: `${p}.sslCert` });

    const comboboxOptions = useMemo(() => {
        return sslCerts.map(cert => ({
            value: { id: cert.id, name: cert.name },
            label: cert.name,
        }));
    }, [sslCerts]);

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
