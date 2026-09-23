import type { HivePaaSRegistrySettings_UpdateOnePayload } from "~/system-settings/api/services";
import type { HivePaaSRegistrySettings } from "~/system-settings/domain";

import type { HivePaaSRegistrySettingsFormInput, HivePaaSRegistrySettingsFormOutput } from "../schemas";

/**
 * What a registry nobody has configured yet starts with.
 *
 * The two numbers are the ones the server defaults to as well: the last ten
 * builds of every app, and everything from the past month. 1gb is the server's
 * default too; it refuses anything below 256mb.
 */
const DEFAULT_KEEP_LAST = 10;
const DEFAULT_KEEP_DAYS = 30;
const DEFAULT_MEMORY_LIMIT = "1gb";

export function toRegistryFormInput(settings?: HivePaaSRegistrySettings): HivePaaSRegistrySettingsFormInput {
    return {
        enabled: settings?.enabled ?? false,
        domain: settings?.domain ?? "",
        storageType: settings?.storage.type ?? "volume",
        volumeId: settings?.storage.volume?.id ?? "",
        cloudStorageId: settings?.storage.cloudStorage?.id ?? "",
        // Off unless it was asked for: zot answers the interface's pages to anybody,
        // and the account that can sign in is the one HivePaaS pushes with.
        dashboardEnabled: settings?.dashboardEnabled ?? false,
        cleanupEnabled: settings?.cleanup.enabled ?? true,
        keepLast: settings?.cleanup.keepLast ?? DEFAULT_KEEP_LAST,
        keepDays: settings?.cleanup.keepDays ?? DEFAULT_KEEP_DAYS,
        memoryLimit: settings?.memoryLimit ?? DEFAULT_MEMORY_LIMIT,
    };
}

/** What a save has to confirm before it may take the registry down. */
export type RegistryRemoval = {
    removeApp: boolean;
    removeStorage: boolean;
};

export function toRegistryPayload(
    values: HivePaaSRegistrySettingsFormOutput,
    updateVer: number,
    removal?: RegistryRemoval,
): HivePaaSRegistrySettings_UpdateOnePayload {
    // Only the store in use is sent. The other id is whatever the form happened
    // to hold before the operator switched, and it names nothing the registry
    // reads.
    const onVolume = values.storageType === "volume";

    return {
        updateVer,
        enabled: values.enabled,
        domain: values.domain,
        storage: {
            type: values.storageType,
            volume: onVolume && values.volumeId ? { id: values.volumeId } : null,
            cloudStorage: !onVolume && values.cloudStorageId ? { id: values.cloudStorageId } : null,
        },
        dashboardEnabled: values.dashboardEnabled,
        cleanup: {
            enabled: values.cleanupEnabled,
            keepLast: values.keepLast,
            keepDays: values.keepDays,
        },
        memoryLimit: values.memoryLimit || DEFAULT_MEMORY_LIMIT,
        removeApp: removal?.removeApp,
        removeStorage: removal?.removeStorage,
    };
}

/** The sentence under the two numbers: a policy nobody can picture from two integers. */
export function describeCleanup(enabled: boolean, keepLast: number, keepDays: number): string {
    if (!enabled) {
        return "Nothing is removed. The registry grows until the disk or the bucket does.";
    }

    const builds = keepLast === 1 ? "build" : "builds";
    const days = keepDays === 1 ? "day" : "days";
    return (
        `Keeps everything from the past ${keepDays} ${days}, and past that the last ` +
        `${keepLast} ${builds} of every app in every environment. ` +
        "Anything else is removed, and its space comes back within a couple of hours."
    );
}
