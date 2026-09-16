import { createContext } from "react";

import {
    AuditLogsApi,
    AuditLogsApiValidator,
    SpecExportApi,
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
        specExport: new SpecExportApi(),
    };

    return {
        operations: operationsApi,
        systemStatus: operationsApi,
    };
}

export const OperationsApiContext = createContext({
    api: createApi(),
});
