import { createContext } from "react";

import {
    SystemBackupApi,
    SystemBackupApiValidator,
    SystemBackupFileApi,
    SystemBackupFileApiValidator,
    SystemCleanupApi,
    SystemCleanupApiValidator,
} from "../services";

function createApi() {
    const systemBackupValidator = new SystemBackupApiValidator();
    const systemBackupFileValidator = new SystemBackupFileApiValidator();
    const systemCleanupValidator = new SystemCleanupApiValidator();

    return {
        systemSettings: {
            backup: new SystemBackupApi(systemBackupValidator),
            backupFiles: new SystemBackupFileApi(systemBackupFileValidator),
            cleanup: new SystemCleanupApi(systemCleanupValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
