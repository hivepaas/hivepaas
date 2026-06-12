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
    SystemSslRenewalApi,
    SystemSslRenewalApiValidator,
    TraefikServiceSettingsApi,
    TraefikServiceSettingsApiValidator,
} from "../services";

function createApi() {
    const systemBackupValidator = new SystemBackupApiValidator();
    const systemBackupFileValidator = new SystemBackupFileApiValidator();
    const systemCleanupValidator = new SystemCleanupApiValidator();
    const systemSslRenewalValidator = new SystemSslRenewalApiValidator();
    const localPaaSServiceSettingsValidator = new LocalPaaSServiceSettingsApiValidator();
    const traefikServiceSettingsValidator = new TraefikServiceSettingsApiValidator();

    return {
        systemSettings: {
            localpaasServiceSettings: new LocalPaaSServiceSettingsApi(localPaaSServiceSettingsValidator),
            traefikServiceSettings: new TraefikServiceSettingsApi(traefikServiceSettingsValidator),
            backup: new SystemBackupApi(systemBackupValidator),
            backupFiles: new SystemBackupFileApi(systemBackupFileValidator),
            cleanup: new SystemCleanupApi(systemCleanupValidator),
            sslRenewal: new SystemSslRenewalApi(systemSslRenewalValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
