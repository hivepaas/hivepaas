import type { HivePaaSServiceSettings } from "~/system-settings/domain";

import { joinTrustedIPsText } from "../hivepaas-general.constants";
import type { HivePaaSGeneralFormInput } from "../schemas";

export function mapHivePaaSServiceSettingsToFormInput(settings: HivePaaSServiceSettings): HivePaaSGeneralFormInput {
    return {
        appSettings: {
            replicas: settings.appSettings.replicas,
        },
        workerSettings: {
            replicas: settings.workerSettings.replicas,
            concurrency: settings.workerSettings.concurrency,
            runWorkerInMainApp: settings.workerSettings.runWorkerInMainApp,
        },
        taskSettings: {
            taskCheckInterval: settings.taskSettings.taskCheckInterval,
            taskCreateInterval: settings.taskSettings.taskCreateInterval,
        },
        periodicSettings: {
            baseInterval: settings.periodicSettings.baseInterval,
            batchSize: settings.periodicSettings.batchSize,
        },
        proxySettings: {
            proxyProvider: settings.proxySettings.proxyProvider,
            trustedIPsText: joinTrustedIPsText(settings.proxySettings.trustedIPs),
        },
    };
}
