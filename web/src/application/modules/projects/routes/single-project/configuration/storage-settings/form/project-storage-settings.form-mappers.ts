import type { ProjectStorageSettings } from "~/projects/domain";

import { type ProjectStorageSettingsFormSchemaInput, emptyProjectStorageSettingsFormDefaults } from "../schemas";

export function mapProjectStorageSettingsToFormInput(
    data: ProjectStorageSettings,
): ProjectStorageSettingsFormSchemaInput {
    return {
        bindSettings: {
            enabled: data.bindSettings?.enabled ?? emptyProjectStorageSettingsFormDefaults.bindSettings.enabled,
            baseDirs: (data.bindSettings?.baseDirs ?? []).map(value => ({ value })),
            subpathTemplate:
                data.bindSettings?.subpathTemplate ??
                emptyProjectStorageSettingsFormDefaults.bindSettings.subpathTemplate,
        },
        volumeSettings: {
            enabled: data.volumeSettings?.enabled ?? emptyProjectStorageSettingsFormDefaults.volumeSettings.enabled,
            volumes: data.volumeSettings?.volumes ?? emptyProjectStorageSettingsFormDefaults.volumeSettings.volumes,
            subpathTemplate:
                data.volumeSettings?.subpathTemplate ??
                emptyProjectStorageSettingsFormDefaults.volumeSettings.subpathTemplate,
        },
        clusterVolumeSettings: {
            enabled:
                data.clusterVolumeSettings?.enabled ??
                emptyProjectStorageSettingsFormDefaults.clusterVolumeSettings.enabled,
            volumes:
                data.clusterVolumeSettings?.volumes ??
                emptyProjectStorageSettingsFormDefaults.clusterVolumeSettings.volumes,
            subpathTemplate:
                data.clusterVolumeSettings?.subpathTemplate ??
                emptyProjectStorageSettingsFormDefaults.clusterVolumeSettings.subpathTemplate,
        },
        tmpfsSettings: {
            enabled: data.tmpfsSettings?.enabled ?? emptyProjectStorageSettingsFormDefaults.tmpfsSettings.enabled,
            maxSize: data.tmpfsSettings?.maxSize ?? emptyProjectStorageSettingsFormDefaults.tmpfsSettings.maxSize,
        },
    };
}
