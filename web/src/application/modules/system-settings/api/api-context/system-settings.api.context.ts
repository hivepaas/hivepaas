import { createContext } from "react";

import {
    HivePaaSAppSecretApi,
    HivePaaSAppSecretApiValidator,
    HivePaaSLoggingSettingsApi,
    HivePaaSLoggingSettingsApiValidator,
    HivePaaSRequestInfoApi,
    HivePaaSRequestInfoApiValidator,
    HivePaaSRestartApi,
    HivePaaSRestartApiValidator,
    HivePaaSRoutingSettingsApi,
    HivePaaSRoutingSettingsApiValidator,
    HivePaaSSecuritySettingsApi,
    HivePaaSSecuritySettingsApiValidator,
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
    TraefikRestartApi,
    TraefikRestartApiValidator,
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
    const hivePaaSSecuritySettingsValidator = new HivePaaSSecuritySettingsApiValidator();
    const hivePaaSAppSecretValidator = new HivePaaSAppSecretApiValidator();
    const hivePaaSLoggingSettingsValidator = new HivePaaSLoggingSettingsApiValidator();
    const hivePaaSRestartValidator = new HivePaaSRestartApiValidator();
    const hivePaaSRequestInfoValidator = new HivePaaSRequestInfoApiValidator();
    const traefikServiceSettingsValidator = new TraefikServiceSettingsApiValidator();
    const traefikConfigOptionsValidator = new TraefikConfigOptionsApiValidator();
    const traefikRestartValidator = new TraefikRestartApiValidator();

    return {
        systemSettings: {
            hivepaasServiceSettings: new HivePaaSServiceSettingsApi(hivePaaSServiceSettingsValidator),
            hivepaasRoutingSettings: new HivePaaSRoutingSettingsApi(hivePaaSRoutingSettingsValidator),
            hivepaasHttpSettings: new HivePaaSRoutingSettingsApi(hivePaaSRoutingSettingsValidator),
            hivepaasSecuritySettings: new HivePaaSSecuritySettingsApi(hivePaaSSecuritySettingsValidator),
            hivepaasAppSecret: new HivePaaSAppSecretApi(hivePaaSAppSecretValidator),
            hivepaasLoggingSettings: new HivePaaSLoggingSettingsApi(hivePaaSLoggingSettingsValidator),
            hivepaasRestart: new HivePaaSRestartApi(hivePaaSRestartValidator),
            hivepaasRequestInfo: new HivePaaSRequestInfoApi(hivePaaSRequestInfoValidator),
            traefikServiceSettings: new TraefikServiceSettingsApi(traefikServiceSettingsValidator),
            traefikConfigOptions: new TraefikConfigOptionsApi(traefikConfigOptionsValidator),
            traefikRestart: new TraefikRestartApi(traefikRestartValidator),
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
