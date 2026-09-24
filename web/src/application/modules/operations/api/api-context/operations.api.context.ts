import { createContext } from "react";

import {
    AuditLogsApi,
    AuditLogsApiValidator,
    SpecExportApi,
    SpecImportApi,
    SpecImportApiMapper,
    SpecImportApiValidator,
    SystemTaskLogsWsApi,
    SystemTasksApi,
    SystemTasksApiValidator,
} from "../services";

function createApi() {
    const systemTasksValidator = new SystemTasksApiValidator();
    const auditLogsValidator = new AuditLogsApiValidator();
    const specImportValidator = new SpecImportApiValidator();
    const specImportMapper = new SpecImportApiMapper();

    const operationsApi = {
        tasks: new SystemTasksApi(systemTasksValidator),
        taskLogs: {
            $: new SystemTaskLogsWsApi(),
        },
        auditLogs: new AuditLogsApi(auditLogsValidator),
        specExport: new SpecExportApi(),
        specImport: new SpecImportApi(specImportValidator, specImportMapper),
    };

    return {
        operations: operationsApi,
        systemStatus: operationsApi,
    };
}

export const OperationsApiContext = createContext({
    api: createApi(),
});
