import { createContext } from "react";

import {
    HivePaaSAppSecretApi,
    HivePaaSAppSecretApiValidator,
    HivePaaSRestartApi,
    HivePaaSRestartApiValidator,
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
    const hivePaaSRestartValidator = new HivePaaSRestartApiValidator();
    const traefikServiceSettingsValidator = new TraefikServiceSettingsApiValidator();
    const traefikConfigOptionsValidator = new TraefikConfigOptionsApiValidator();

    return {
        systemSettings: {
            hivepaasServiceSettings: new HivePaaSServiceSettingsApi(hivePaaSServiceSettingsValidator),
            hivepaasRoutingSettings: new HivePaaSRoutingSettingsApi(hivePaaSRoutingSettingsValidator),
            hivepaasHttpSettings: new HivePaaSRoutingSettingsApi(hivePaaSRoutingSettingsValidator),
            hivepaasAppSecret: new HivePaaSAppSecretApi(hivePaaSAppSecretValidator),
            hivepaasRestart: new HivePaaSRestartApi(hivePaaSRestartValidator),
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
