import { createContext } from "react";

import { SystemTaskLogsWsApi, SystemTasksApi, SystemTasksApiValidator } from "../services";

function createApi() {
    const systemTasksValidator = new SystemTasksApiValidator();

    return {
        systemStatus: {
            tasks: new SystemTasksApi(systemTasksValidator),
            taskLogs: {
                $: new SystemTaskLogsWsApi(),
            },
        },
    };
}

export const SystemStatusApiContext = createContext({
    api: createApi(),
});
