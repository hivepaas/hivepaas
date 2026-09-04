import { createContext } from "react";

import {
    HivePaaSAppSecretApi,
    HivePaaSAppSecretApiValidator,
    HivePaaSRoutingSettingsApi,
    HivePaaSRoutingSettingsApiValidator,
    HivePaaSServiceSettingsApi,
    HivePaaSServiceSettingsApiValidator,
    SystemBackupApi,
    SystemBackupApiValidator,
    SystemBackupFileApi,
    SystemBackupFileApiValidator,
    SystemBackupRepoCleanupApi,
    SystemBackupRepoCleanupApiValidator,
    SystemCleanupApi,
    SystemCleanupApiValidator,
    SystemSslRenewalApi,
    SystemSslRenewalApiValidator,
    TraefikConfigOptionsApi,
    TraefikConfigOptionsApiValidator,
    TraefikServiceSettingsApi,
    TraefikServiceSettingsApiValidator,
} from "../services";

function createApi() {
    const systemBackupValidator = new SystemBackupApiValidator();
    const systemBackupFileValidator = new SystemBackupFileApiValidator();
    const systemCleanupValidator = new SystemCleanupApiValidator();
    const systemSslRenewalValidator = new SystemSslRenewalApiValidator();
    const systemBackupRepoCleanupValidator = new SystemBackupRepoCleanupApiValidator();
    const hivePaaSServiceSettingsValidator = new HivePaaSServiceSettingsApiValidator();
    const hivePaaSRoutingSettingsValidator = new HivePaaSRoutingSettingsApiValidator();
    const hivePaaSAppSecretValidator = new HivePaaSAppSecretApiValidator();
    const traefikServiceSettingsValidator = new TraefikServiceSettingsApiValidator();
    const traefikConfigOptionsValidator = new TraefikConfigOptionsApiValidator();

    return {
        systemSettings: {
            hivepaasServiceSettings: new HivePaaSServiceSettingsApi(hivePaaSServiceSettingsValidator),
            hivepaasRoutingSettings: new HivePaaSRoutingSettingsApi(hivePaaSRoutingSettingsValidator),
            hivepaasHttpSettings: new HivePaaSRoutingSettingsApi(hivePaaSRoutingSettingsValidator),
            hivepaasAppSecret: new HivePaaSAppSecretApi(hivePaaSAppSecretValidator),
            traefikServiceSettings: new TraefikServiceSettingsApi(traefikServiceSettingsValidator),
            traefikConfigOptions: new TraefikConfigOptionsApi(traefikConfigOptionsValidator),
            backup: new SystemBackupApi(systemBackupValidator),
            backupFiles: new SystemBackupFileApi(systemBackupFileValidator),
            cleanup: new SystemCleanupApi(systemCleanupValidator),
            sslRenewal: new SystemSslRenewalApi(systemSslRenewalValidator),
            backupRepoCleanup: new SystemBackupRepoCleanupApi(systemBackupRepoCleanupValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
