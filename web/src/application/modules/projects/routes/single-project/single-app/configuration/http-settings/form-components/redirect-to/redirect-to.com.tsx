import React, { useMemo } from "react";

import { FieldError } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";

import { EditableCombobox, InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

function View({ domainIndex }: RedirectToProps) {
    const { control, setValue } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const domains = useWatch({ control, name: "domains" });
    const currentDomain = useWatch({ control, name: `domains.${domainIndex}.domain` });

    const {
        field: domainRedirect,
        fieldState: { error: domainRedirectError, invalid: isDomainRedirectInvalid },
    } = useController({ control, name: `domains.${domainIndex}.domainRedirect` });

    const options = useMemo(() => {
        const deduped = new Set<string>();
        for (const item of domains as ({ domain?: string } | undefined)[]) {
            const rawDomain = item?.domain;
            const value = typeof rawDomain === "string" ? rawDomain.trim() : "";
            if (!value || value === currentDomain) {
                continue;
            }
            deduped.add(value);
        }
        return Array.from(deduped);
    }, [currentDomain, domains]);

    return (
        <InfoBlock title="Redirect To">
            <EditableCombobox
                options={options}
                value={domainRedirect.value}
                onChange={value => {
                    setValue(`domains.${domainIndex}.domainRedirect`, value, { shouldDirty: true });
                }}
                placeholder="https://other-domain.com"
                className="max-w-[400px]"
                allowClear
                aria-invalid={isDomainRedirectInvalid}
            />
            <FieldError errors={[domainRedirectError]} />
        </InfoBlock>
    );
}

interface RedirectToProps {
    domainIndex: number;
}

export const RedirectTo = React.memo(View);
