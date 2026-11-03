import type { PropsWithChildren } from "react";

import { ProfileQueries } from "@application/shared/data/queries";
import { EProfileApiKeyStatus } from "@application/shared/enums";

/** The first page of the person's keys: a key beyond it is reported as not found, which is only less helpful. */
const KEYS_LISTED = { pagination: { page: 1, size: 100 } };

interface Props {
    keyId: string;
    /** Whether assistants may make changes, as saved. */
    allowWrite: boolean;
}

/**
 * What the pasted key will let an assistant do, said before it is put into a
 * client: a key that may only read still only reads with changes allowed, and
 * nothing on the server says so until somebody asks it for a change.
 *
 * The key is looked up among the person's own keys by its ID - the secret is
 * not sent anywhere. The list is read once, not per keystroke.
 */
export function McpKeyCheck({ keyId, allowWrite }: Props) {
    const { data, isLoading } = ProfileQueries.useFindManyApiKeysPaginated(KEYS_LISTED, {
        enabled: keyId !== "",
    });
    if (keyId === "" || isLoading || !data) {
        return null;
    }

    const found = data.data.find(apiKey => apiKey.keyId === keyId);
    if (!found) {
        return (
            <Message tone="info">
                This key is not among yours here, so this page cannot say what it may do. An assistant acts as the
                key&apos;s user, with the key&apos;s access.
            </Message>
        );
    }
    if (found.status !== EProfileApiKeyStatus.Active) {
        return <Message tone="warning">This key is not active: the server will refuse it.</Message>;
    }
    if (found.expireAt && found.expireAt.getTime() < Date.now()) {
        return (
            <Message tone="warning">
                This key expired on {found.expireAt.toLocaleDateString()}: the server will refuse it.
            </Message>
        );
    }
    if (!allowWrite || !found.accessAction) {
        // Changes are off, so any key only reads; or the key is not limited.
        return null;
    }

    const { execute, write } = found.accessAction;
    if (!execute && !write) {
        return (
            <Message tone="warning">
                This key may only read. With changes allowed, an assistant using it still only reads: it will say it
                needs a key with Execute to restart or redeploy apps, or Write to install them, change their
                configuration or schedule jobs.
            </Message>
        );
    }
    if (!write) {
        return (
            <Message tone="info">
                This key may restart and redeploy apps. Installing, changing configuration and scheduling jobs need
                Write.
            </Message>
        );
    }
    if (!execute) {
        return (
            <Message tone="info">
                This key may install apps, change configuration and schedule jobs. Restarting and redeploying need
                Execute.
            </Message>
        );
    }
    return null;
}

function Message({ tone, children }: PropsWithChildren<{ tone: "info" | "warning" }>) {
    return (
        <p
            className={
                tone === "warning" ? "text-xs text-amber-700 dark:text-amber-400" : "text-xs text-muted-foreground"
            }
        >
            {children}
        </p>
    );
}
