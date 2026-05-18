import { createContext } from "react";

import { SystemBackupApi, SystemBackupApiValidator, SystemCleanupApi, SystemCleanupApiValidator } from "../services";

function createApi() {
    const systemBackupValidator = new SystemBackupApiValidator();
    const systemCleanupValidator = new SystemCleanupApiValidator();

    return {
        systemSettings: {
            backup: new SystemBackupApi(systemBackupValidator),
            cleanup: new SystemCleanupApi(systemCleanupValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
