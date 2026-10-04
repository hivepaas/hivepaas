import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OperationsApiContext } from "~/operations/api/api-context";
import type { ComposeImport_Apply_Req, ComposeImport_Validate_Req } from "~/operations/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { isSpecImportErrorHandledByCaller } from "./spec-import.errors";

function createHook() {
    return function useComposeImportApi() {
        const { api } = use(OperationsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                // Validate runs again on every change of the file or of a choice,
                // and what it refuses is shown beside the file rather than as a
                // popup each time.
                validate: async (data: ComposeImport_Validate_Req["data"], signal?: AbortSignal) => {
                    const result = await api.operations.composeImport.validate({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                apply: async (data: ComposeImport_Apply_Req["data"], signal?: AbortSignal) => {
                    const result = await api.operations.composeImport.apply({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isSpecImportErrorHandledByCaller(error)) {
                                notifyError({ message: "Failed to create the project", error });
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

export const useComposeImportApi = createHook();
