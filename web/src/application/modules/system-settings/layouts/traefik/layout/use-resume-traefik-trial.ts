import { useEffect } from "react";

import { TraefikConfigOptionsQueries } from "~/system-settings/data";
import { useSettingsChangeConfirmDialogState } from "~/system-settings/dialogs";

/**
 * Reopens a config options trial that is still running, from any page under
 * System > Traefik.
 *
 * This matters more here than anywhere else. Applying config options replaces
 * Traefik's task, and its published ports are bound in host mode - so the old
 * task stops before the new one can bind, and the HTTP response to the change
 * that started the trial is sent down a connection that is being cut. It often
 * never arrives, which means the dialog that would have opened from it never
 * opens either. Without this the operator's own change expires while they are
 * looking at the page that started it.
 *
 * Watching from the layout rather than from the config options route means
 * landing on any Traefik page is enough, including after the reload somebody
 * naturally reaches for when the tab has gone blank.
 *
 * The dialog store refuses a trial it has already seen through to an end, so this
 * cannot resurrect one from a query result that has not caught up yet.
 */
export function useResumeTraefikTrial(): void {
    const dialog = useSettingsChangeConfirmDialogState();
    const configOptions = TraefikConfigOptionsQueries.useFindOne();

    const pending = configOptions.data?.data.pendingChange ?? null;

    useEffect(() => {
        if (pending != null) {
            dialog.open("traefik", pending);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pending]);
}
