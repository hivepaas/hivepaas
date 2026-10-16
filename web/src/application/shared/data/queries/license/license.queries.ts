import { type UseQueryOptions, useQuery } from "@tanstack/react-query";

import { useLicenseApi } from "@application/shared/api";
import type { License_FindOne_Res } from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";

/**
 * The license in force, with its usage
 */
type FindOneRes = License_FindOne_Res;

type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

/**
 * Silent when it fails: see useLicenseApi. No placeholder data either - it is one
 * object, and a previous one would only show a plan no longer in force.
 *
 * Only for someone who can read System: anyone else gets a 403. Outside System,
 * use useReadableLicense, which asks only then.
 */
function useFindOne(options: FindOneOptions = {}) {
    const { queries } = useLicenseApi();

    return useQuery({
        queryKey: [QK["license.find-one"]],
        queryFn: ({ signal }) => queries.findOne(signal),

        ...options,
    });
}

export const LicenseQueries = Object.freeze({
    useFindOne,
});
