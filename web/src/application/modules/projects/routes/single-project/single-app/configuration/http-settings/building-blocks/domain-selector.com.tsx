import { useEffect, useMemo, useState } from "react";
import { Button, Checkbox } from "@components/ui";
import { Badge } from "@components/ui/badge";
import { Plus } from "lucide-react";
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
    const [draft, setDraft] = useState("");
    const [duplicateDomainError, setDuplicateDomainError] = useState<string | null>(null);

    const activeDomainValue = domainValues[activeDomainIndex]?.domain ?? "";
    const normalizeDomain = useMemo(
        () => (value: string) => value.trim().toLowerCase(),
        [],
    );

    useEffect(() => {
        setDraft(activeDomainValue);
    }, [activeDomainIndex, activeDomainValue]);

    function handleDomainSelect(value: string) {
        const selectedIndex = domainValues.findIndex(domain => normalizeDomain(domain.domain) === normalizeDomain(value));
        if (selectedIndex >= 0) {
            setActiveDomainIndex(selectedIndex);
            setDraft(domainValues[selectedIndex]?.domain ?? value);
        } else {
            setDraft(value);
        }
        setDuplicateDomainError(null);
    }

    function handleAddDomain() {
        const trimmedDraft = draft.trim();
        if (trimmedDraft.length > 0) {
            const hasDuplicateDomain = domainValues.some(domain => normalizeDomain(domain.domain) === normalizeDomain(trimmedDraft));
            if (hasDuplicateDomain) {
                setDuplicateDomainError("Domain already exists.");
                return;
            }
            append({ ...emptyDomain, domain: trimmedDraft });
            setActiveDomainIndex(fields.length);
            setDraft(trimmedDraft);
            setDuplicateDomainError(null);
            return;
        }

        append({ ...emptyDomain });
        setActiveDomainIndex(fields.length);
        setDraft("");
        setDuplicateDomainError(null);
    }

    return (
        <div className="flex flex-col gap-4">
            {internalEndpoints.length > 0 && (
                <InfoBlock title="Internal Endpoints">
                    <div className="flex flex-wrap gap-2">
                        {internalEndpoints.map(ep => (
                            <Badge
                                key={ep}
                                variant="outline"
                                className="font-mono text-xs"
                            >
                                {ep}
                            </Badge>
                        ))}
                    </div>
                </InfoBlock>
            )}

            {domainSuggestion && (
                <InfoBlock title="Domain Suggestion">
                    <span className="text-sm text-muted-foreground font-mono">{domainSuggestion}</span>
                </InfoBlock>
            )}

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Expose Publicly"
                        content="Allow external access to this app via HTTP/HTTPS"
                    />
                }
            >
                <Checkbox
                    checked={exposePublicly.value}
                    onCheckedChange={exposePublicly.onChange}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Domain"
                        content="Select an existing domain to edit or type a new domain name"
                    />
                }
            >
                <div className="flex items-center gap-2 max-w-[500px]">
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
                        className="flex-1"
                    />
                    <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        title="Add domain"
                        onClick={handleAddDomain}
                    >
                        <Plus className="size-4" />
                    </Button>
                </div>
                {duplicateDomainError ? <p className="text-destructive text-sm">{duplicateDomainError}</p> : null}
            </InfoBlock>
        </div>
    );
}
