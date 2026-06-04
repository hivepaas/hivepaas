import type { LocalPaaSServiceSettings } from "~/system-settings/domain";

import type { LocalPaaSGeneralFormInput } from "../schemas";

export function mapLocalPaaSServiceSettingsToFormInput(settings: LocalPaaSServiceSettings): LocalPaaSGeneralFormInput {
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
