import { useState } from "react";

import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "@/components/ui";

import { CopyField } from "./copy-field.com";

/**
 * How each MCP client is pointed at the server with an API key. The server is
 * the same for all of them - MCP is not any one vendor's - but each client
 * writes its configuration its own way, and some cannot send a key at all.
 *
 * Checked against each client's documentation in September 2026. A client that
 * connects to remote servers with OAuth only - ChatGPT's and claude.ai's
 * connectors - is named as such rather than given a configuration that fails.
 */

interface Step {
    /** What to do with the value, in a sentence. */
    say: string;
    /** What the toast calls it once copied. */
    what: string;
    value: string;
}

interface ClientGuide {
    value: string;
    label: string;
    steps: Step[];
    notes?: string[];
}

/** The environment variable the clients that read one take the key from. */
const TOKEN_ENV = "HIVEPAAS_MCP_TOKEN";

function clientGuides(endpoint: string, keyId: string, secret: string): ClientGuide[] {
    // The server takes the key as its ID and secret joined by a colon, as a
    // bearer token, from any client that can send only Authorization.
    const token = `${keyId}:${secret}`;
    return [
        {
            value: "claude-code",
            label: "Claude Code",
            steps: [
                {
                    say: "Run in a terminal:",
                    what: "Command",
                    value:
                        `claude mcp add --transport http hivepaas ${endpoint} \\\n` +
                        `  --header "HIVEPAAS-API-KEY-ID: ${keyId}" \\\n` +
                        `  --header "HIVEPAAS-API-SECRET-KEY: ${secret}"`,
                },
            ],
            notes: ["Add --scope user to have it in every project. /mcp in Claude Code shows whether it connected."],
        },
        {
            value: "claude-desktop",
            label: "Claude Desktop",
            steps: [
                {
                    say: "Add to claude_desktop_config.json (Settings › Developer › Edit Config), then restart Claude Desktop:",
                    what: "Configuration",
                    value: JSON.stringify(
                        {
                            mcpServers: {
                                hivepaas: {
                                    command: "npx",
                                    args: [
                                        "-y",
                                        "mcp-remote",
                                        endpoint,
                                        "--transport",
                                        "http-only",
                                        "--header",
                                        // Literal: mcp-remote expands it. No space around the colon: some
                                        // clients pass arguments with spaces to npx broken.
                                        "Authorization:${AUTH_HEADER}",
                                    ],
                                    env: { AUTH_HEADER: `Bearer ${token}` },
                                },
                            },
                        },
                        null,
                        2,
                    ),
                },
            ],
            notes: [
                "Claude Desktop does not connect to a remote server listed in its configuration file, and its Connectors take OAuth only. mcp-remote is the bridge that sends the key; it needs Node.js.",
            ],
        },
        {
            value: "codex",
            label: "Codex CLI",
            steps: [
                {
                    say: "Keep the key in your shell's environment - in ~/.zshrc or ~/.bashrc:",
                    what: "Line",
                    value: `export ${TOKEN_ENV}="${token}"`,
                },
                {
                    say: "Then add the server:",
                    what: "Command",
                    value: `codex mcp add hivepaas --url ${endpoint} --bearer-token-env-var ${TOKEN_ENV}`,
                },
            ],
            notes: ["OpenAI's Codex. Start it from a shell that has the variable; codex mcp list shows the server."],
        },
        {
            value: "gemini",
            label: "Gemini CLI",
            steps: [
                {
                    say: "Run in a terminal:",
                    what: "Command",
                    value: `gemini mcp add --transport http --header "Authorization: Bearer ${token}" hivepaas ${endpoint}`,
                },
            ],
            notes: ["Google's Gemini CLI. /mcp in it shows whether it connected."],
        },
        {
            value: "vscode",
            label: "VS Code (Copilot)",
            steps: [
                {
                    say: "Add to .vscode/mcp.json, or to the file MCP: Open User Configuration opens for every workspace:",
                    what: "Configuration",
                    value: JSON.stringify(
                        {
                            inputs: [
                                {
                                    type: "promptString",
                                    id: "hivepaas-key",
                                    description: "HivePaaS API key, as <key-id>:<secret>",
                                    password: true,
                                },
                            ],
                            servers: {
                                hivepaas: {
                                    type: "http",
                                    url: endpoint,
                                    headers: { Authorization: "Bearer ${input:hivepaas-key}" },
                                },
                            },
                        },
                        null,
                        2,
                    ),
                },
                {
                    say: "VS Code asks for the key the first time the server starts, and keeps it out of the file. Paste:",
                    what: "Key",
                    value: token,
                },
            ],
        },
        {
            value: "cursor",
            label: "Cursor",
            steps: [
                {
                    say: "Add to ~/.cursor/mcp.json, or to .cursor/mcp.json in a project:",
                    what: "Configuration",
                    value: JSON.stringify(
                        { mcpServers: { hivepaas: { url: endpoint, headers: { Authorization: `Bearer ${token}` } } } },
                        null,
                        2,
                    ),
                },
            ],
            notes: ["The key is in the file as it is: keep a project's .cursor/mcp.json out of version control."],
        },
        {
            value: "other",
            label: "Other clients",
            steps: [
                {
                    say: "Any client that speaks MCP over streamable HTTP and sends a header connects with:",
                    what: "Header",
                    value: `Authorization: Bearer ${token}`,
                },
            ],
            notes: [
                "Or the two headers HIVEPAAS-API-KEY-ID and HIVEPAAS-API-SECRET-KEY.",
                "ChatGPT's and claude.ai's connectors connect to remote servers with OAuth only, which HivePaaS does not offer yet: use one of the clients above instead.",
            ],
        },
    ];
}

interface Props {
    endpoint: string;
    keyId: string;
    secret: string;
}

/**
 * One picker, two faces: seven tabs fit in a row only when the space beside
 * the row's title is wide, which depends on the sidebars as much as on the
 * screen - so the switch is a container query on this block, not a breakpoint
 * of the viewport. Narrower than that, the same choice is a dropdown.
 */
export function McpClientSnippets({ endpoint, keyId, secret }: Props) {
    const guides = clientGuides(endpoint, keyId, secret);
    const [client, setClient] = useState("claude-code");

    return (
        <div className="@container flex min-w-0 flex-col gap-2">
            <Select
                value={client}
                onValueChange={setClient}
            >
                <SelectTrigger
                    className="w-full @3xl:hidden"
                    aria-label="Client"
                >
                    <SelectValue />
                </SelectTrigger>
                <SelectContent>
                    {guides.map(guide => (
                        <SelectItem
                            key={guide.value}
                            value={guide.value}
                        >
                            {guide.label}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
            <Tabs
                value={client}
                onValueChange={setClient}
            >
                <TabsList className="hidden @3xl:inline-flex">
                    {guides.map(guide => (
                        <TabsTrigger
                            key={guide.value}
                            value={guide.value}
                        >
                            {guide.label}
                        </TabsTrigger>
                    ))}
                </TabsList>
                {guides.map(guide => (
                    <TabsContent
                        key={guide.value}
                        value={guide.value}
                        className="flex flex-col gap-3 pt-2"
                    >
                        {guide.steps.map(step => (
                            <div
                                key={step.say}
                                className="flex flex-col gap-1"
                            >
                                <span className="text-sm">{step.say}</span>
                                <CopyField
                                    what={step.what}
                                    value={step.value}
                                />
                            </div>
                        ))}
                        {guide.notes?.map(note => (
                            <p
                                key={note}
                                className="text-xs text-muted-foreground"
                            >
                                {note}
                            </p>
                        ))}
                    </TabsContent>
                ))}
            </Tabs>
        </div>
    );
}
