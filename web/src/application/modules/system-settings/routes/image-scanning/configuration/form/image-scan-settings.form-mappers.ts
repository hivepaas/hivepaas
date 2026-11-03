import type { SystemImageScan_UpdateOne_Payload } from "~/system-settings/api/services";
import type { SystemImageScanSettings } from "~/system-settings/domain";

import type { ImageScanSettingsFormInput, ImageScanSettingsFormOutput } from "../schemas";

const DEFAULT_CPU_LIMIT = 1;
const DEFAULT_MEMORY_LIMIT = "512mb";

export function toImageScanFormInput(settings?: SystemImageScanSettings): ImageScanSettingsFormInput {
    return {
        enabled: settings?.enabled ?? false,
        volumeId: settings?.volume?.id ?? "",
        cpuLimit: settings?.cpuLimit ?? DEFAULT_CPU_LIMIT,
        memoryLimit: settings?.memoryLimit ?? DEFAULT_MEMORY_LIMIT,
        dbRepository: settings?.dbRepository ?? "",
        javaDbRepository: settings?.javaDbRepository ?? "",
        scheduleInterval: settings?.schedule.interval ?? "1d",
        scheduleFrom: settings?.schedule.initialTime ?? null,
        notification: {
            successUseDefault: settings?.notification?.successUseDefault ?? true,
            success: settings?.notification?.success,
            failureUseDefault: settings?.notification?.failureUseDefault ?? true,
            failure: settings?.notification?.failure,
        },
    };
}

export function toImageScanPayload(
    values: ImageScanSettingsFormOutput,
    settings?: SystemImageScanSettings,
): SystemImageScan_UpdateOne_Payload {
    return {
        updateVer: settings?.updateVer ?? 0,
        enabled: values.enabled,
        volume: values.volumeId ? { id: values.volumeId } : null,
        dbRepository: values.dbRepository,
        javaDbRepository: values.javaDbRepository,
        schedule: {
            interval: values.scheduleInterval,
            ...(values.scheduleFrom ? { initialTime: values.scheduleFrom } : {}),
        },
        notification: {
            successUseDefault: values.notification.successUseDefault,
            success: { id: values.notification.successUseDefault ? "" : (values.notification.success?.id ?? "") },
            failureUseDefault: values.notification.failureUseDefault,
            failure: { id: values.notification.failureUseDefault ? "" : (values.notification.failure?.id ?? "") },
        },
        cpuLimit: values.cpuLimit ?? DEFAULT_CPU_LIMIT,
        memoryLimit: values.memoryLimit || DEFAULT_MEMORY_LIMIT,
    };
}
