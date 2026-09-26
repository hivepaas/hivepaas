import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { KeyRound } from "lucide-react";
import { Link } from "react-router";
import { toast } from "sonner";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ProfileCommands } from "@application/shared/data/commands";

import { Button, Input } from "@/components/ui";
import { PasswordInput } from "@/components/ui/input-password";

import { CopyField } from "./copy-field.com";

interface CreatedKey {
    keyId: string;
    secretKey: string;
}

/** Placeholders in the snippets until a key is pasted or created here. */
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

/** Which access actions a key needs, so the person can create the one they want in their profile. */
function KeyAccessNote({ allowWrite }: { allowWrite: boolean }) {
    return (
        <ul className="list-disc pl-5 text-xs leading-relaxed text-muted-foreground">
            <li>Reading - apps, logs, tasks, the store - needs Read.</li>
            {allowWrite ? (
                <>
                    <li>Restarting and redeploying an app needs Execute.</li>
                    <li>Installing, changing configuration and scheduling jobs needs Write.</li>
                    <li>
                        A key that can make changes is created in{" "}
                        <Link
                            to={ROUTE.currentUser.profileApiKeys.create.$route}
                            className="underline underline-offset-2"
                        >
                            Profile › API keys
                        </Link>
                        , with the access actions it needs and no more.
                    </li>
                </>
            ) : (
                <li>Changes are off: whatever the key may do, an assistant only reads.</li>
            )}
        </ul>
    );
}

function todayName(): string {
    return `MCP - ${new Date().toISOString().slice(0, 10)}`;
}

interface Props {
    endpoint: string;
    /** Whether the server is on as saved: a key made while it is off answers 404 until it is not. */
    enabled: boolean;
    /** Whether assistants may make changes, as saved. */
    allowWrite: boolean;
}

/**
 * How to point an assistant at the server: a key, and the configuration each
 * kind of client takes, filled in with it.
 *
 * The key is the person's own: pasted here, or - for one that may only read -
 * created here. A key that can make changes is only ever created in the
 * profile, where its access actions are chosen deliberately. What is pasted
 * stays in this page: it fills the snippets and is neither saved nor sent.
 */
export function McpConnectSection({ endpoint, enabled, allowWrite }: Props) {
    const [key, setKey] = useState<CreatedKey>({ keyId: "", secretKey: "" });
    const [wasCreated, setWasCreated] = useState(false);
    const { mutate: createApiKey, isPending } = ProfileCommands.useCreateOneApiKey();

    function handleCreate() {
        createApiKey(
            {
                name: todayName(),
                accessAction: { read: true, write: false, execute: false, delete: false },
            },
            {
                onSuccess: response => {
                    setKey({ keyId: response.data.keyId, secretKey: response.data.secretKey });
                    setWasCreated(true);
                    toast.success("Read-only API key created");
                },
            },
        );
    }

    const keyId = key.keyId.trim() || KEY_ID_PLACEHOLDER;
    const secret = key.secretKey.trim() || SECRET_PLACEHOLDER;

    return (
        <>
            <SectionHeader>Connect a client</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="API key"
                            content="The server takes an API key and nothing else, and acts as the key's user, within the key's access actions. Paste one here to fill in the snippets below; it stays in this page and is not saved."
                        />
                    }
                >
                    <div className="flex max-w-[640px] flex-col gap-2">
                        <Input
                            placeholder="Key ID"
                            autoComplete="off"
                            value={key.keyId}
                            onChange={event => {
                                setKey(prev => ({ ...prev, keyId: event.target.value }));
                            }}
                        />
                        <PasswordInput
                            placeholder="Secret"
                            autoComplete="off"
                            value={key.secretKey}
                            onChange={event => {
                                setKey(prev => ({ ...prev, secretKey: event.target.value }));
                            }}
                        />
                        {wasCreated ? (
                            <div className={cn(dashedBorderBox, "text-sm")}>
                                <span className="font-semibold text-orange-500">Copy it now:</span> the secret is not
                                shown again. The snippets below carry it. Manage the key in your profile, under API
                                keys.
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
                                {!enabled && (
                                    <span className="text-xs text-muted-foreground">
                                        The server answers once it is enabled and saved.
                                    </span>
                                )}
                            </div>
                        )}
                        <KeyAccessNote allowWrite={allowWrite} />
                    </div>
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
