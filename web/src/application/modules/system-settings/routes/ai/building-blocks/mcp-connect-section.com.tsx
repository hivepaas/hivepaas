import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ExternalLink, KeyRound } from "lucide-react";
import { Link } from "react-router";
import { toast } from "sonner";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ProfileCommands } from "@application/shared/data/commands";

import { Button, Input } from "@/components/ui";
import {
    AlertDialog,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { PasswordInput } from "@/components/ui/input-password";

import { McpClientSnippets } from "./mcp-client-snippets.com";
import { McpKeyCheck } from "./mcp-key-check.com";

interface CreatedKey {
    keyId: string;
    secretKey: string;
}

interface KeyState extends CreatedKey {
    /** Set for a key created here, which the page says when it runs out. */
    expireAt?: Date;
}

/** Placeholders in the snippets until a key is pasted or created here. */
const KEY_ID_PLACEHOLDER = "<key-id>";
const SECRET_PLACEHOLDER = "<secret>";

/** Which access actions each kind of tool needs, for a key made in the profile. */
function KeyAccessNote({ allowWrite }: { allowWrite: boolean }) {
    return (
        <ul className="list-disc pl-5 text-xs leading-relaxed text-muted-foreground">
            <li>Reading - apps, logs, tasks, the store - needs Read.</li>
            {allowWrite ? (
                <>
                    <li>Restarting and redeploying an app needs Execute.</li>
                    <li>Installing, changing configuration and scheduling jobs needs Write.</li>
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

/**
 * The key the button makes: what the server lets an assistant do, and no more.
 * HivePaaS requires every key to expire within a year; a key pasted into a
 * client's configuration is one nobody looks at again, so it runs out well
 * before that - and the sooner, the more it may do.
 */
interface KeyKind {
    label: string;
    created: string;
    access: { read: boolean; execute: boolean; write: boolean; delete: boolean };
    lifetimeDays: number;
    /** Whether the person confirms first: a key that can change things works on the whole API. */
    confirm: boolean;
}

const READ_ONLY_KEY: KeyKind = {
    label: "Create a read-only key",
    created: "Read-only API key created",
    access: { read: true, execute: false, write: false, delete: false },
    lifetimeDays: 90,
    confirm: false,
};

const CHANGES_KEY: KeyKind = {
    label: "Create a key that can make changes",
    created: "API key that can make changes created",
    access: { read: true, execute: true, write: true, delete: false },
    lifetimeDays: 30,
    confirm: true,
};

function keyExpiry(days: number): Date {
    const expireAt = new Date();
    expireAt.setDate(expireAt.getDate() + days);
    return expireAt;
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
 * The key is the person's own: pasted here, made in the profile, or made by the
 * button - which makes the key the server's setting calls for, read-only or one
 * that can make changes, the latter only once the person has confirmed what it
 * can do. What is pasted stays in this page: it fills the snippets and is
 * neither saved nor sent.
 */
export function McpConnectSection({ endpoint, enabled, allowWrite }: Props) {
    const [key, setKey] = useState<KeyState>({ keyId: "", secretKey: "" });
    const [wasCreated, setWasCreated] = useState(false);
    const [isConfirming, setIsConfirming] = useState(false);
    const { mutate: createApiKey, isPending } = ProfileCommands.useCreateOneApiKey();
    const kind = allowWrite ? CHANGES_KEY : READ_ONLY_KEY;

    function create() {
        const expireAt = keyExpiry(kind.lifetimeDays);
        createApiKey(
            { name: todayName(), accessAction: kind.access, expireAt },
            {
                onSuccess: response => {
                    setKey({ keyId: response.data.keyId, secretKey: response.data.secretKey, expireAt });
                    setWasCreated(true);
                    setIsConfirming(false);
                    toast.success(kind.created);
                },
            },
        );
    }

    function handleCreate() {
        if (kind.confirm) {
            setIsConfirming(true);
            return;
        }
        create();
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
                                shown again. The snippets below carry it. The key runs out on{" "}
                                {key.expireAt?.toLocaleDateString()}; manage it in your profile, under API keys.
                            </div>
                        ) : (
                            <div className="flex flex-col items-start gap-2">
                                <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
                                    <Button
                                        type="button"
                                        variant="outline"
                                        onClick={handleCreate}
                                        isLoading={isPending && !isConfirming}
                                        disabled={isPending}
                                    >
                                        <KeyRound className="size-4" />
                                        {kind.label}
                                    </Button>
                                    <Link
                                        to={ROUTE.currentUser.profileApiKeys.$route}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="inline-flex items-center gap-1 text-sm underline underline-offset-2"
                                    >
                                        Create one yourself
                                        <ExternalLink className="size-3.5" />
                                    </Link>
                                </div>
                                <span className="text-xs text-muted-foreground">
                                    It lasts {kind.lifetimeDays} days. Your own, in your profile, can have other access
                                    actions and another lifetime.
                                </span>
                                {!enabled && (
                                    <span className="text-xs text-muted-foreground">
                                        The server answers once it is enabled and saved.
                                    </span>
                                )}
                            </div>
                        )}
                        <McpKeyCheck
                            keyId={key.keyId.trim()}
                            allowWrite={allowWrite}
                        />
                        <KeyAccessNote allowWrite={allowWrite} />
                    </div>
                    <ConfirmChangesKeyDialog
                        open={isConfirming}
                        lifetimeDays={CHANGES_KEY.lifetimeDays}
                        isPending={isPending}
                        onOpenChange={setIsConfirming}
                        onConfirm={create}
                    />
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Client"
                            content="MCP is not any one vendor's: every client below reaches the same server, and whichever model it runs sees the same tools. Each writes its configuration its own way."
                        />
                    }
                >
                    <McpClientSnippets
                        endpoint={endpoint}
                        keyId={keyId}
                        secret={secret}
                    />
                </InfoBlock>
            </div>
        </>
    );
}

interface ConfirmProps {
    open: boolean;
    lifetimeDays: number;
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void;
}

/**
 * What a key that can make changes can do, said before it exists: it acts on
 * the whole API, not only through the MCP server's plans.
 */
function ConfirmChangesKeyDialog({ open, lifetimeDays, isPending, onOpenChange, onConfirm }: ConfirmProps) {
    return (
        <AlertDialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <AlertDialogContent className="sm:max-w-[520px]">
                <AlertDialogHeader>
                    <AlertDialogTitle>Create a key that can make changes?</AlertDialogTitle>
                    <AlertDialogDescription asChild>
                        <ul className="list-disc space-y-2 pl-5 text-left text-sm">
                            <li>
                                It may read, execute and write, and not delete. An assistant using it can restart and
                                redeploy apps, install them, change their configuration and schedule jobs - each once
                                you agree to its plan.
                            </li>
                            <li className="text-amber-700 dark:text-amber-400">
                                It works on the whole HivePaaS API, not only through the MCP server: whoever holds it
                                can do all of that directly, with no plan, as you. Keep it out of shared and versioned
                                files.
                            </li>
                            <li>It runs out in {lifetimeDays} days, and you can revoke it in your profile.</li>
                        </ul>
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
                    <Button
                        type="button"
                        onClick={onConfirm}
                        isLoading={isPending}
                        disabled={isPending}
                    >
                        Create the key
                    </Button>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}
