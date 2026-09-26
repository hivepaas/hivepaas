import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { KeyRound } from "lucide-react";
import { toast } from "sonner";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ProfileCommands } from "@application/shared/data/commands";

import { Button } from "@/components/ui";

import { CopyField } from "./copy-field.com";

interface CreatedKey {
    keyId: string;
    secretKey: string;
}

/** Placeholders in the snippets until a key is created here. */
const KEY_ID_PLACEHOLDER = "<key-id>";
const SECRET_PLACEHOLDER = "<secret>";

function claudeCodeSnippet(endpoint: string, keyId: string, secret: string): string {
    return (
        `claude mcp add --transport http hivepaas ${endpoint} \\\n` +
        `  --header "HIVEPAAS-API-KEY-ID: ${keyId}" \\\n` +
        `  --header "HIVEPAAS-API-SECRET-KEY: ${secret}"`
    );
}

function mcpServersSnippet(endpoint: string, keyId: string, secret: string): string {
    const config = {
        mcpServers: {
            hivepaas: {
                type: "http",
                url: endpoint,
                headers: {
                    "HIVEPAAS-API-KEY-ID": keyId,
                    "HIVEPAAS-API-SECRET-KEY": secret,
                },
            },
        },
    };
    return JSON.stringify(config, null, 2);
}

function todayName(): string {
    return `MCP - ${new Date().toISOString().slice(0, 10)}`;
}

interface Props {
    endpoint: string;
    /** Whether the server is on as saved: a key made while it is off answers 404 until it is not. */
    enabled: boolean;
}

/**
 * How to point an assistant at the server: a key that may only read, and the
 * configuration each kind of client takes, filled in with that key the moment
 * it is created. The secret is shown this once, as everywhere a key is made.
 */
export function McpConnectSection({ endpoint, enabled }: Props) {
    const [createdKey, setCreatedKey] = useState<CreatedKey | null>(null);
    const { mutate: createApiKey, isPending } = ProfileCommands.useCreateOneApiKey();

    function handleCreate() {
        createApiKey(
            {
                name: todayName(),
                accessAction: { read: true, write: false, execute: false, delete: false },
            },
            {
                onSuccess: response => {
                    setCreatedKey({ keyId: response.data.keyId, secretKey: response.data.secretKey });
                    toast.success("Read-only API key created");
                },
            },
        );
    }

    const keyId = createdKey?.keyId ?? KEY_ID_PLACEHOLDER;
    const secret = createdKey?.secretKey ?? SECRET_PLACEHOLDER;

    return (
        <>
            <SectionHeader>Connect a client</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="API key"
                            content="The server takes an API key and nothing else. A key limited to reading lets an assistant see what you can see, and change nothing. It is yours: the assistant sees only what you may."
                        />
                    }
                >
                    {createdKey ? (
                        <div className="flex flex-col gap-2">
                            <div className={cn(dashedBorderBox, "text-sm")}>
                                <span className="font-semibold text-orange-500">Copy it now:</span> the secret is not
                                shown again. The snippets below carry it. Manage the key in your profile, under API
                                keys.
                            </div>
                            <CopyField
                                what="Key ID"
                                value={createdKey.keyId}
                            />
                            <CopyField
                                what="Secret"
                                value={createdKey.secretKey}
                            />
                        </div>
                    ) : (
                        <div className="flex flex-col items-start gap-2">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={handleCreate}
                                isLoading={isPending}
                                disabled={isPending}
                            >
                                <KeyRound className="size-4" />
                                Create a read-only key
                            </Button>
                            <span className="text-xs text-muted-foreground">
                                Or use a key you have, in the snippets below.
                                {!enabled && " The server answers once it is enabled and saved."}
                            </span>
                        </div>
                    )}
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Claude Code"
                            content="Run it in a terminal. Add --scope user to have it in every project."
                        />
                    }
                >
                    <CopyField
                        what="Command"
                        value={claudeCodeSnippet(endpoint, keyId, secret)}
                    />
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Claude Desktop and editors"
                            content="The mcpServers entry that Claude Desktop, Cursor, VS Code and most other clients read from their configuration file."
                        />
                    }
                >
                    <CopyField
                        what="Configuration"
                        value={mcpServersSnippet(endpoint, keyId, secret)}
                    />
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Other clients"
                            content="A client that sets only the Authorization header can send the key there, as its id and secret joined by a colon."
                        />
                    }
                >
                    <CopyField
                        what="Header"
                        value={`Authorization: Bearer ${keyId}:${secret}`}
                    />
                </InfoBlock>
            </div>
        </>
    );
}
