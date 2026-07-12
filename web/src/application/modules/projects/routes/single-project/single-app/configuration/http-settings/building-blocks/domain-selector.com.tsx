import { useEffect, useMemo, useState } from "react";

import { Button, Checkbox, FieldError } from "@components/ui";
import { EyeIcon, Plus } from "lucide-react";
import { useController, useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { isValidDomain } from "@application/shared/utils/domain";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    emptyDomain,
} from "../schemas";

interface DomainSelectorProps {
    activeDomainIndex: number;
    setActiveDomainIndex: (index: number) => void;
    internalEndpoints: string[];
    domainSuggestion: string;
    readOnly?: boolean;
}

export function DomainSelector({
    activeDomainIndex,
    setActiveDomainIndex,
    internalEndpoints,
    domainSuggestion,
    readOnly = false,
}: DomainSelectorProps) {
    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const { append } = useFieldArray({ control, name: "domains" });

    const { field: exposePublicly } = useController({ control, name: "exposePublicly" });

    const domainValues = useWatch({ control, name: "domains", defaultValue: [] });
    const normalizeDomain = useMemo(() => (value: string) => value.trim().toLowerCase(), []);
    const suggestedDomain = domainSuggestion.trim();
    const domainNames = useMemo(() => {
        const configuredDomains = domainValues.map(d => d.domain).filter(domain => domain.trim().length > 0);

        if (
            suggestedDomain.length === 0 ||
            configuredDomains.some(domain => normalizeDomain(domain) === normalizeDomain(suggestedDomain))
        ) {
            return configuredDomains;
        }

        return [...configuredDomains, suggestedDomain];
    }, [domainValues, normalizeDomain, suggestedDomain]);
    const [draft, setDraft] = useState(suggestedDomain);
    const [domainInputError, setDomainInputError] = useState<string | null>(null);

    const {
        fieldState: { error: activeDomainFieldError },
    } = useController({
        control,
        name: `domains.${activeDomainIndex >= 0 ? activeDomainIndex : 0}.domain`,
    });

    const activeDomain = activeDomainIndex >= 0 ? domainValues[activeDomainIndex] : undefined;
    const hasActiveDomain = Boolean(activeDomain);
    const activeDomainValue = activeDomain?.domain ?? "";
    const selectedDomain = hasActiveDomain ? activeDomainValue.trim() : "";
    const canViewSelectedDomain = selectedDomain.length > 0;

    useEffect(() => {
        setDraft(hasActiveDomain ? activeDomainValue : suggestedDomain);
    }, [activeDomainValue, hasActiveDomain, suggestedDomain]);

    function handleDomainSelect(value: string) {
        if (readOnly) {
            return;
        }

        const selectedIndex = domainValues.findIndex(
            domain => normalizeDomain(domain.domain) === normalizeDomain(value),
        );
        if (selectedIndex >= 0) {
            setActiveDomainIndex(selectedIndex);
            setDraft(domainValues[selectedIndex]?.domain ?? value);
        } else {
            setDraft(value);
        }
        setDomainInputError(null);
    }

    function handleAddDomain() {
        if (readOnly) {
            return;
        }

        const firstPort = domainValues[0]?.containerPort;
        const containerPort =
            typeof firstPort === "number" && Number.isInteger(firstPort) && firstPort >= 1 && firstPort <= 65535
                ? firstPort
                : emptyDomain.containerPort;

        const trimmedDraft = draft.trim();
        if (trimmedDraft.length > 0) {
            if (!isValidDomain(trimmedDraft, { maxLength: 100 })) {
                setDomainInputError("Enter a valid domain (e.g. app.example.com)");
                return;
            }
            const hasDuplicateDomain = domainValues.some(
                domain => normalizeDomain(domain.domain) === normalizeDomain(trimmedDraft),
            );
            if (hasDuplicateDomain) {
                setDomainInputError("Domain already exists.");
                return;
            }
            append({ ...emptyDomain, domain: trimmedDraft, containerPort });
            setActiveDomainIndex(domainValues.length);
            setDraft(trimmedDraft);
            setDomainInputError(null);
            return;
        }

        append({ ...emptyDomain, containerPort });
        setActiveDomainIndex(domainValues.length);
        setDraft("");
        setDomainInputError(null);
    }

    return (
        <div className="flex flex-col gap-4">
            <InfoBlock title="Project Internal Endpoint(s)">
                <div className="flex flex-wrap gap-2">
                    {internalEndpoints.map(ep => (
                        <span
                            key={ep}
                            className="font-mono text-xs text-red-500"
                        >
                            {ep}
                        </span>
                    ))}
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Expose The App To The Internet"
                        content="Allow external access to this app via HTTP/HTTPS"
                    />
                }
            >
                <Checkbox
                    checked={exposePublicly.value}
                    onCheckedChange={value => {
                        if (readOnly) {
                            return;
                        }

                        exposePublicly.onChange(value);
                    }}
                    disabled={readOnly}
                />
            </InfoBlock>
            {exposePublicly.value && (
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Domain"
                            content="Select an existing domain to edit or type a new domain name"
                        />
                    }
                >
                    <div className="flex items-center gap-2 ">
                        <EditableCombobox
                            options={domainNames}
                            value={draft}
                            onInputChange={(value: string) => {
                                if (readOnly) {
                                    return;
                                }

                                setDraft(value);
                                setDomainInputError(null);
                            }}
                            onChange={handleDomainSelect}
                            onRefresh={undefined}
                            placeholder="e.g. app.example.com"
                            emptyText="No domains configured"
                            className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                            disableFilter
                            disabled={readOnly}
                        />
                        <div className="w-[74px]">
                            <Button
                                type="button"
                                variant="outline"
                                title="Add domain"
                                onClick={handleAddDomain}
                                disabled={readOnly}
                            >
                                <Plus className="size-4" /> Add
                            </Button>
                        </div>
                        {canViewSelectedDomain && (
                            <Button
                                type="button"
                                variant="outline"
                                title="View domain"
                                onClick={() => {
                                    const url = /^https?:\/\//i.test(selectedDomain)
                                        ? selectedDomain
                                        : `https://${selectedDomain}`;
                                    window.open(url, "_blank", "noopener,noreferrer");
                                }}
                            >
                                <EyeIcon className="size-4" /> View
                            </Button>
                        )}
                    </div>
                    <FieldError errors={[domainInputError ? { message: domainInputError } : activeDomainFieldError]} />
                </InfoBlock>
            )}
        </div>
    );
}
