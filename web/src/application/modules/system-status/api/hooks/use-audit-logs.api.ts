import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemStatusApiContext } from "~/system-status/api/api-context";
import type {
    AuditLogs_FindManyPaginated_Req,
    AuditLogs_FindOneById_Req,
    AuditLogs_FindTypes_Req,
} from "~/system-status/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAuditLogsApi() {
        const { api } = use(SystemStatusApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: AuditLogs_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemStatus.auditLogs.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get audit logs", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: AuditLogs_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemStatus.auditLogs.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get audit log", error });
                            throw error;
                        },
                    });
                },
                findTypes: async (data: AuditLogs_FindTypes_Req["data"] = {}, signal?: AbortSignal) => {
                    const result = await api.systemStatus.auditLogs.findTypes({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: () => ({ data: [] as string[] }),
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            queries,
        };
    };
}

export const useAuditLogsApi = createHook();
