import { createContext } from "react";

import {
    HivePaaSAppSecretApi,
    HivePaaSAppSecretApiValidator,
    HivePaaSLoggingPerformanceApi,
    HivePaaSLoggingPerformanceApiValidator,
    HivePaaSLoggingSettingsApi,
    HivePaaSLoggingSettingsApiValidator,
    HivePaaSRegistrySettingsApi,
    HivePaaSRegistrySettingsApiValidator,
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
    HivePaaSUpdatesApi,
    HivePaaSUpdatesApiValidator,
    McpSettingsApi,
    McpSettingsApiValidator,
    SystemBackupApi,
    SystemBackupApiValidator,
    SystemBackupRepoCleanupApi,
    SystemBackupRepoCleanupApiValidator,
    SystemCleanupApi,
    SystemCleanupApiValidator,
    SystemRegistryAuthRenewalApi,
    SystemRegistryAuthRenewalApiValidator,
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
    const systemCleanupValidator = new SystemCleanupApiValidator();
    const systemSslRenewalValidator = new SystemSslRenewalApiValidator();
    const systemRegistryAuthRenewalValidator = new SystemRegistryAuthRenewalApiValidator();
    const systemBackupRepoCleanupValidator = new SystemBackupRepoCleanupApiValidator();
    const hivePaaSServiceSettingsValidator = new HivePaaSServiceSettingsApiValidator();
    const hivePaaSRoutingSettingsValidator = new HivePaaSRoutingSettingsApiValidator();
    const hivePaaSSecuritySettingsValidator = new HivePaaSSecuritySettingsApiValidator();
    const hivePaaSAppSecretValidator = new HivePaaSAppSecretApiValidator();
    const hivePaaSLoggingSettingsValidator = new HivePaaSLoggingSettingsApiValidator();
    const hivePaaSRegistrySettingsValidator = new HivePaaSRegistrySettingsApiValidator();
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
            hivepaasLoggingPerformance: new HivePaaSLoggingPerformanceApi(new HivePaaSLoggingPerformanceApiValidator()),
            hivepaasRegistrySettings: new HivePaaSRegistrySettingsApi(hivePaaSRegistrySettingsValidator),
            hivepaasRestart: new HivePaaSRestartApi(hivePaaSRestartValidator),
            hivepaasUpdates: new HivePaaSUpdatesApi(new HivePaaSUpdatesApiValidator()),
            hivepaasRequestInfo: new HivePaaSRequestInfoApi(hivePaaSRequestInfoValidator),
            traefikServiceSettings: new TraefikServiceSettingsApi(traefikServiceSettingsValidator),
            traefikConfigOptions: new TraefikConfigOptionsApi(traefikConfigOptionsValidator),
            traefikRestart: new TraefikRestartApi(traefikRestartValidator),
            backup: new SystemBackupApi(systemBackupValidator),
            cleanup: new SystemCleanupApi(systemCleanupValidator),
            sslRenewal: new SystemSslRenewalApi(systemSslRenewalValidator),
            registryAuthRenewal: new SystemRegistryAuthRenewalApi(systemRegistryAuthRenewalValidator),
            backupRepoCleanup: new SystemBackupRepoCleanupApi(systemBackupRepoCleanupValidator),
            mcpSettings: new McpSettingsApi(new McpSettingsApiValidator()),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
