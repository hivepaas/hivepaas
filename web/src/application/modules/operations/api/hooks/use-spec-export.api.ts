import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OperationsApiContext } from "~/operations/api/api-context";
import type { SpecExport_Export_Req } from "~/operations/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSpecExportApi() {
        const { api } = use(OperationsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                exportSpec: async (data: SpecExport_Export_Req["data"], signal?: AbortSignal) => {
                    const result = await api.operations.specExport.exportSpec({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to export the configuration spec", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { mutations };
    };
}

export const useSpecExportApi = createHook();
