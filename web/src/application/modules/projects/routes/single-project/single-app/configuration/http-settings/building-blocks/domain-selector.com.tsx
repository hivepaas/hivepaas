import { useEffect, useMemo, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { EyeIcon, Plus } from "lucide-react";
import { useController, useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

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
}

export function DomainSelector({
    activeDomainIndex,
    setActiveDomainIndex,
    internalEndpoints,
    domainSuggestion,
}: DomainSelectorProps) {
    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const { fields, append } = useFieldArray({ control, name: "domains" });

    const { field: exposePublicly } = useController({ control, name: "exposePublicly" });

    const domainValues = useWatch({ control, name: "domains", defaultValue: [] });
    const domainNames = domainValues.map(d => d.domain).filter(Boolean);
    const [draft, setDraft] = useState(domainSuggestion);
    const [duplicateDomainError, setDuplicateDomainError] = useState<string | null>(null);

    const activeDomainValue = domainValues[activeDomainIndex]?.domain ?? "";
    const selectedDomain = activeDomainIndex >= 0 ? activeDomainValue.trim() : "";
    const canViewSelectedDomain = selectedDomain.length > 0;
    const normalizeDomain = useMemo(() => (value: string) => value.trim().toLowerCase(), []);

    useEffect(() => {
        setDraft(activeDomainValue);
    }, [activeDomainIndex, activeDomainValue]);

    useEffect(() => {
        setDraft(domainSuggestion);
    }, [domainSuggestion]);

    function handleDomainSelect(value: string) {
        const selectedIndex = domainValues.findIndex(
            domain => normalizeDomain(domain.domain) === normalizeDomain(value),
        );
        if (selectedIndex >= 0) {
            setActiveDomainIndex(selectedIndex);
            setDraft(domainValues[selectedIndex]?.domain ?? value);
        } else {
            setDraft(value);
        }
        setDuplicateDomainError(null);
    }

    function handleAddDomain() {
        const firstPort = domainValues[0]?.containerPort;
        const containerPort =
            typeof firstPort === "number" && Number.isInteger(firstPort) && firstPort >= 1 && firstPort <= 65535
                ? firstPort
                : emptyDomain.containerPort;

        const trimmedDraft = draft.trim();
        if (trimmedDraft.length > 0) {
            const hasDuplicateDomain = domainValues.some(
                domain => normalizeDomain(domain.domain) === normalizeDomain(trimmedDraft),
            );
            if (hasDuplicateDomain) {
                setDuplicateDomainError("Domain already exists.");
                return;
            }
            append({ ...emptyDomain, domain: trimmedDraft, containerPort });
            setActiveDomainIndex(fields.length);
            setDraft(trimmedDraft);
            setDuplicateDomainError(null);
            return;
        }

        append({ ...emptyDomain, containerPort });
        setActiveDomainIndex(fields.length);
        setDraft("");
        setDuplicateDomainError(null);
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
                    onCheckedChange={exposePublicly.onChange}
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
                                setDraft(value);
                                setDuplicateDomainError(null);
                            }}
                            onChange={handleDomainSelect}
                            onRefresh={undefined}
                            placeholder="e.g. app.example.com"
                            emptyText="No domains configured"
                            className="max-w-[400px]"
                            disableFilter
                        />
                        <div className="w-[74px]">
                            <Button
                                type="button"
                                variant="outline"
                                title="Add domain"
                                onClick={handleAddDomain}
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
                    {duplicateDomainError ? <p className="text-destructive text-sm">{duplicateDomainError}</p> : null}
                </InfoBlock>
            )}
        </div>
    );
}
