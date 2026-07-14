import type { HivePaaSServiceSettings } from "~/system-settings/domain";

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
        healthcheckSettings: {
            baseInterval: settings.healthcheckSettings.baseInterval,
        },
    };
}
