import type { UseQueryOptions, UseQueryResult } from "@tanstack/react-query";

import type { License_FindOne_Res } from "@application/shared/api/services";
import { MODULE_IDS } from "@application/shared/constants";
import { LicenseQueries } from "@application/shared/data/queries";
import type { License } from "@application/shared/entities";
import { useConditionalModule } from "@application/shared/permissions/hooks";

type ReadableLicenseOptions = Omit<UseQueryOptions<License_FindOne_Res>, "queryKey" | "queryFn">;

export interface ReadableLicense {
    /** Whether this person can read System, which the license belongs to. */
    canRead: boolean;
    /** The license in force. Undefined while it loads, when it could not be read, and for anyone who cannot read System. */
    license: License | undefined;
    /** The query, for its loading and error states. It never runs for anyone who cannot read System. */
    query: UseQueryResult<License_FindOne_Res>;
}

/**
 * The license, for the people who may read it: those with read access to System.
 * For anyone else it is never asked for - the server would refuse with a 403 - and
 * is undefined, so whatever reads it shows and marks nothing.
 *
 * This is how the license is read outside System › License: the banners, the
 * plan-limit dialog and the OAuth page. Like the query, it fails silently.
 */
export function useReadableLicense(options: ReadableLicenseOptions = {}): ReadableLicense {
    const { canRead } = useConditionalModule({ id: MODULE_IDS.System });
    const query = LicenseQueries.useFindOne({ ...options, enabled: canRead && (options.enabled ?? true) });

    return { canRead, license: canRead ? query.data?.data : undefined, query };
}
