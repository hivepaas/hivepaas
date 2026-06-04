import { createContext } from "react";

import {
    LocalPaaSServiceSettingsApi,
    LocalPaaSServiceSettingsApiValidator,
    SystemBackupApi,
    SystemBackupApiValidator,
    SystemBackupFileApi,
    SystemBackupFileApiValidator,
    SystemCleanupApi,
    SystemCleanupApiValidator,
    TraefikServiceSettingsApi,
    TraefikServiceSettingsApiValidator,
} from "../services";

function createApi() {
    const systemBackupValidator = new SystemBackupApiValidator();
    const systemBackupFileValidator = new SystemBackupFileApiValidator();
    const systemCleanupValidator = new SystemCleanupApiValidator();
    const localPaaSServiceSettingsValidator = new LocalPaaSServiceSettingsApiValidator();
    const traefikServiceSettingsValidator = new TraefikServiceSettingsApiValidator();

    return {
        systemSettings: {
            localpaasServiceSettings: new LocalPaaSServiceSettingsApi(localPaaSServiceSettingsValidator),
            traefikServiceSettings: new TraefikServiceSettingsApi(traefikServiceSettingsValidator),
            backup: new SystemBackupApi(systemBackupValidator),
            backupFiles: new SystemBackupFileApi(systemBackupFileValidator),
            cleanup: new SystemCleanupApi(systemCleanupValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
