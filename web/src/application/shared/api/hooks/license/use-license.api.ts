import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";
import type { License_InstallKey_Req } from "@application/shared/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { isLicenseKeyRefusal } from "./license.errors";

function createHook() {
    return function useLicenseApi() {
        const { api } = use(ApplicationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Deliberately silent about failures: it throws without a toast.
                 *
                 * It runs on every page, for the license banners, where a failure
                 * is nothing the person asked about - a toast there would follow
                 * them from page to page. System › License shows its own error in
                 * the page instead.
                 */
                findOne: async (signal?: AbortSignal) => {
                    const result = await api.license.findOne({ data: {} }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                installKey: async (data: License_InstallKey_Req["data"]) => {
                    const result = await api.license.installKey({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            // A key the server refuses is answered beside the box it
                            // was pasted into: see isLicenseKeyRefusal.
                            if (!isLicenseKeyRefusal(error)) {
                                notifyError({ message: "Failed to install the license key", error });
                            }

                            throw error;
                        },
                    });
                },
                refresh: async () => {
                    const result = await api.license.refresh({ data: {} });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Could not check for a newer key", error });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useLicenseApi = createHook();
