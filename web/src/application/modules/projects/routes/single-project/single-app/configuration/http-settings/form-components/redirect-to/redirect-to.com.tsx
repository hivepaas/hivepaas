import React, { useMemo } from "react";

import { FieldError } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";

import { EditableCombobox, InfoBlock } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

function View({ domainIndex, readOnly = false }: RedirectToProps) {
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
                    if (readOnly) {
                        return;
                    }

                    setValue(`domains.${domainIndex}.domainRedirect`, value, { shouldDirty: true });
                }}
                placeholder="e.g. other-domain.com"
                className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                allowClear
                aria-invalid={isDomainRedirectInvalid}
                disabled={readOnly}
            />
            <FieldError errors={[domainRedirectError]} />
        </InfoBlock>
    );
}

interface RedirectToProps {
    domainIndex: number;
    readOnly?: boolean;
}

export const RedirectTo = React.memo(View);
