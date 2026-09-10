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

    const operationsApi = {
        tasks: new SystemTasksApi(systemTasksValidator),
        taskLogs: {
            $: new SystemTaskLogsWsApi(),
        },
        auditLogs: new AuditLogsApi(auditLogsValidator),
    };

    return {
        operations: operationsApi,
        systemStatus: operationsApi,
    };
}

export const OperationsApiContext = createContext({
    api: createApi(),
});
