import { createContext } from "react";

import {
    AuditLogsApi,
    AuditLogsApiValidator,
    SystemTaskLogsWsApi,
    SystemTasksApi,
    SystemTasksApiValidator,
} from "../services";

function createApi() {
    const systemTasksValidator = new SystemTasksApiValidator();
    const auditLogsValidator = new AuditLogsApiValidator();

    return {
        systemStatus: {
            tasks: new SystemTasksApi(systemTasksValidator),
            taskLogs: {
                $: new SystemTaskLogsWsApi(),
            },
            auditLogs: new AuditLogsApi(auditLogsValidator),
        },
    };
}

export const SystemStatusApiContext = createContext({
    api: createApi(),
});
