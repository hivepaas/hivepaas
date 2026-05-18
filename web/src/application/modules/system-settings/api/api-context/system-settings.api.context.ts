import { createContext } from "react";

import { SystemBackupApi, SystemBackupApiValidator } from "../services";

function createApi() {
    const systemBackupValidator = new SystemBackupApiValidator();

    return {
        systemSettings: {
            backup: new SystemBackupApi(systemBackupValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
