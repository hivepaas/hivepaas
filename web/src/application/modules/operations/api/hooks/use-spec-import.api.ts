import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OperationsApiContext } from "~/operations/api/api-context";
import type { SpecImport_Apply_Req, SpecImport_Validate_Req } from "~/operations/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { isSpecImportErrorHandledByCaller } from "./spec-import.errors";

function createHook() {
    return function useSpecImportApi() {
        const { api } = use(OperationsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                // Validate runs again on every change of the selection, and a
                // bundle it cannot read is shown beside the file rather than as
                // a popup each time.
                validate: async (data: SpecImport_Validate_Req["data"], signal?: AbortSignal) => {
                    const result = await api.operations.specImport.validate({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                apply: async (data: SpecImport_Apply_Req["data"], signal?: AbortSignal) => {
                    const result = await api.operations.specImport.apply({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isSpecImportErrorHandledByCaller(error)) {
                                notifyError({ message: "Failed to import the configuration spec", error });
                            }
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

export const useSpecImportApi = createHook();
