import { ArrowUp } from "lucide-react";
import { Link } from "react-router";
import { HivePaaSUpdatesQueries } from "~/system-settings/data";

import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";

/**
 * The version HivePaaS runs, and the one it can move to when there is one.
 *
 * Only for admins: updating is theirs to do, and the server tells nobody else
 * what is published. Stable is named first when both channels have something
 * newer; a beta is named only when it is all there is.
 */
export function UpdateAvailableBadge() {
    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const { data } = HivePaaSUpdatesQueries.useFindReleaseInfo({ enabled: isAdmin, retry: false }, { quiet: true });

    if (!isAdmin || !data) {
        return null;
    }

    const { current, stable, beta } = data.data;
    const target = stable?.canUpdate ? stable : beta?.canUpdate ? beta : null;
    const isBeta = target !== null && target === beta;

    if (!target) {
        return current ? (
            <span className="text-[13px] text-muted-foreground">
                HivePaaS <span className="font-mono">{current.appVersion}</span>
            </span>
        ) : null;
    }

    return (
        <Link
            to={ROUTE.systemSettings.hivepaas.updates.$route}
            className="inline-flex h-9 items-center gap-2 rounded-lg border border-amber-300 bg-amber-50 px-3 text-[13px] font-medium text-amber-950 hover:bg-amber-100 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-100 dark:hover:bg-amber-900"
        >
            <ArrowUp className="size-3.5" />
            {current && <span className="font-mono">{current.appVersion}</span>}
            <span aria-hidden>·</span>
            <span>
                Update to <span className="font-mono">{target.appVersion}</span>
                {isBeta && " (beta)"}
            </span>
        </Link>
    );
}
