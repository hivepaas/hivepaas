import { useEffect } from "react";

import { HivePaaSRoutingSettingsQueries, HivePaaSServiceSettingsQueries } from "~/system-settings/data";
import { useSettingsChangeConfirmDialogState } from "~/system-settings/dialogs";

/**
 * Reopens a trial that is still running, from any page under System > HivePaaS.
 *
 * A change stays on trial whether or not the operator is still on the page that
 * started it, and reloading is exactly what somebody does when a proxy change has
 * just restarted traefik and the tab has gone blank. Watching from the layout
 * rather than from the two settings routes means coming back to any of these four
 * pages is enough, and the operator does not have to remember which one it was.
 *
 * It watches from here rather than from the module's dialog container because the
 * container also covers Traefik, backups, cleanup and SSL renewal - pages where no
 * trial can ever exist, and where these two requests would be pure waste. The
 * dialog itself stays in the container, so a trial already on screen survives a
 * walk into those pages.
 *
 * The dialog store refuses a trial it has already seen through to an end, so this
 * cannot resurrect one from a query result that has not caught up yet.
 */
export function useResumeSettingsTrial(): void {
    const dialog = useSettingsChangeConfirmDialogState();
    const routingSettings = HivePaaSRoutingSettingsQueries.useFindOne();
    const serviceSettings = HivePaaSServiceSettingsQueries.useFindOne();

    const routingPending = routingSettings.data?.data.pendingChange ?? null;
    const servicePending = serviceSettings.data?.data.pendingChange ?? null;

    useEffect(() => {
        if (routingPending != null) {
            dialog.open("routing", routingPending);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [routingPending]);

    useEffect(() => {
        if (servicePending != null) {
            dialog.open("service", servicePending);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [servicePending]);
}
