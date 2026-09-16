import { useState } from "react";

import {
    Button,
    Field,
    FieldGroup,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@components/ui";
import { PasswordInput } from "@components/ui/input-password";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangleIcon } from "lucide-react";
import { SpecExportCommands } from "~/operations/data";
import type { SpecExportResult, SpecExportScope, SpecSecretsMode } from "~/operations/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { type ModuleId, PermissionTooltipAction } from "@application/shared/permissions";

import { SpecExportResultPanel } from "./spec-export-result.com";

const SECRETS_MODE_OPTIONS: { value: SpecSecretsMode; label: string; hint: string }[] = [
    {
        value: "omit",
        label: "Exclude secrets",
        hint: "Secret fields are emptied, so the bundle is safe to share or commit. The keys stay, so you can see which ones need filling in.",
    },
    {
        value: "encrypted",
        label: "Include, encrypted",
        hint: "Secrets travel in the bundle, and the whole archive is encrypted with a passphrase you choose. It opens with the standard `age` tool.",
    },
    {
        value: "plaintext",
        label: "Include, in plain text",
        hint: "Secrets travel readable by anyone who opens the file. Use this only when you are moving an installation and control where the file goes.",
    },
];

/**
 * Downloading a blob needs an anchor: the response is already in memory, and
 * there is no URL a browser could navigate to that would carry the auth header.
 */
function saveBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");

    anchor.href = url;
    anchor.download = filename;
    document.body.append(anchor);
    anchor.click();
    anchor.remove();
    window.URL.revokeObjectURL(url);
}

export function SpecExportPanel({ scope, scopeLabel, permissionModuleId = MODULE_IDS.System }: Props) {
    const [secretsMode, setSecretsMode] = useState<SpecSecretsMode>("omit");
    const [passphrase, setPassphrase] = useState("");
    const [acknowledgedPlaintext, setAcknowledgedPlaintext] = useState(false);
    const [result, setResult] = useState<SpecExportResult | undefined>();

    const { mutate: exportSpec, isPending } = SpecExportCommands.useExportSpec({
        onSuccess: response => {
            saveBlob(response.data.blob, response.data.filename);
            setResult({
                filename: response.data.filename,
                sizeBytes: response.data.sizeBytes,
                summary: response.data.summary,
            });
        },
    });

    const needsPassphrase = secretsMode === "encrypted";
    const needsAcknowledgement = secretsMode === "plaintext";
    const canExport =
        !isPending && (!needsPassphrase || passphrase.length > 0) && (!needsAcknowledgement || acknowledgedPlaintext);

    const selectedHint = SECRETS_MODE_OPTIONS.find(option => option.value === secretsMode)?.hint;

    return (
        <div className="flex flex-col gap-6">
            <div className="rounded-lg border bg-background p-4">
                <div className="flex flex-col items-start gap-6">
                    <p className="text-sm font-medium text-foreground">
                        Export the configuration of {scopeLabel} as a spec - a single archive of YAML describing
                        settings, apps and how they are deployed. Everything below the chosen scope is included.
                    </p>

                    <div className="w-full">
                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Secrets"
                                    content="Values that HivePaaS stores encrypted - passwords, tokens, private keys. The stored form is readable only by this installation, so a bundle never carries it as-is."
                                />
                            }
                        >
                            <FieldGroup>
                                <Field>
                                    <Select
                                        value={secretsMode}
                                        onValueChange={value => {
                                            setSecretsMode(value as SpecSecretsMode);
                                            setAcknowledgedPlaintext(false);
                                        }}
                                    >
                                        <SelectTrigger className="max-w-[280px]">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {SECRETS_MODE_OPTIONS.map(option => (
                                                <SelectItem
                                                    key={option.value}
                                                    value={option.value}
                                                >
                                                    {option.label}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    {selectedHint && (
                                        <p className="mt-2 max-w-[560px] text-xs text-muted-foreground">
                                            {selectedHint}
                                        </p>
                                    )}
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </div>

                    {needsPassphrase && (
                        <div className="w-full">
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Passphrase"
                                        content="Used to encrypt the archive. HivePaaS does not store it - lose it and the bundle cannot be opened."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <PasswordInput
                                            value={passphrase}
                                            onChange={event => {
                                                setPassphrase(event.target.value);
                                            }}
                                            placeholder="Required"
                                            className="max-w-[280px]"
                                        />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </div>
                    )}

                    {needsAcknowledgement && (
                        <div className={cn(dashedBorderBox, "w-full")}>
                            <label className="flex items-start gap-3 text-sm">
                                <input
                                    type="checkbox"
                                    className="mt-1"
                                    checked={acknowledgedPlaintext}
                                    onChange={event => {
                                        setAcknowledgedPlaintext(event.target.checked);
                                    }}
                                />
                                <span>
                                    <AlertTriangleIcon className="mr-1 inline size-4 text-orange-500" />I understand
                                    that every secret in this scope will be readable by anyone who opens the downloaded
                                    file.
                                </span>
                            </label>
                        </div>
                    )}

                    <PermissionTooltipAction
                        id={permissionModuleId}
                        action="read"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="button"
                                className="min-w-[120px]"
                                disabled={!canExport || isDenied}
                                isLoading={isPending}
                                onClick={() => {
                                    if (isDenied) {
                                        return;
                                    }
                                    exportSpec({
                                        scope,
                                        secretsMode,
                                        passphrase: needsPassphrase ? passphrase : undefined,
                                    });
                                }}
                            >
                                Export
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </div>
            </div>

            {result && <SpecExportResultPanel result={result} />}
        </div>
    );
}

interface Props {
    scope: SpecExportScope;
    /** Read in a sentence: "the configuration of {scopeLabel}". */
    scopeLabel: string;
    permissionModuleId?: ModuleId;
}
