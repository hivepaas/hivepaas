import { toast } from "sonner";

import { useGlobalAlertDialogState } from "@application/shared/dialogs/global-alert/hooks/use-global-alert.dialog.state";

import { type MetaNotifications } from "@infrastructure/api/utils";

/**
 * Dedupe window in milliseconds.
 * Prevents refetch/polling from stacking identical notifications.
 */
const DEDUPE_WINDOW_MS = 3000;

const seenMessages = new Map<string, number>();

function isDuplicate(key: string): boolean {
    const now = Date.now();

    // Prune stale entries
    for (const [k, ts] of seenMessages.entries()) {
        if (now - ts > DEDUPE_WINDOW_MS) {
            seenMessages.delete(k);
        }
    }

    if (seenMessages.has(key)) {
        return true;
    }

    seenMessages.set(key, now);

    return false;
}

/**
 * Imperatively notify the user about meta.message, meta.warning, and meta.error
 * fields returned by the backend. Designed to be called from a plain module
 * (e.g. an Axios interceptor) without React context.
 *
 * - message  → sonner info toast (stable id prevents stacking)
 * - warning  → GlobalAlertDialog (type: "error", title: "Warning")
 * - error    → GlobalAlertDialog (type: "error", title: "Error")
 * - both warning + error → single modal, error takes precedence for title
 */
export function notifyApiMeta(meta: MetaNotifications): void {
    if (meta.message) {
        const toastKey = `message:${meta.message}`;

        if (!isDuplicate(toastKey)) {
            toast.info(meta.message, { id: toastKey });
        }
    }

    if (meta.warning || meta.error) {
        let title: string;
        let description: string;

        if (meta.error && meta.warning) {
            title = "Error";
            description = `${meta.error}\n\n${meta.warning}`;
        } else if (meta.error) {
            title = "Error";
            description = meta.error;
        } else {
            title = "Warning";
            description = meta.warning!;
        }

        const modalKey = `modal:${description}`;

        if (isDuplicate(modalKey)) {
            return;
        }

        // Also skip if the dialog is already showing the same description
        const dialogState = useGlobalAlertDialogState.getState();

        if (dialogState.mode === "open" && dialogState.props.description === description) {
            return;
        }

        useGlobalAlertDialogState.getState().open({
            props: {
                title,
                description,
                showFooter: false,
                type: "error",
            },
        });
    }
}
