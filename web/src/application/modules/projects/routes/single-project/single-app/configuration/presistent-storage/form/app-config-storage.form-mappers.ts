import { type AppStorageSettings } from "~/projects/domain";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import { type AppConfigStorageFormSchemaInput, type AppConfigStorageFormSchemaOutput } from "../schemas";

export function mapAppStorageSettingsToFormInput(settings: AppStorageSettings): AppConfigStorageFormSchemaInput {
    return {
        mounts:
            settings.mounts?.map(item => ({
                type: item.type ?? EMountType.Bind,
                source: item.source ?? "",
                target: item.target ?? "",
                consistency: item.consistency ?? EMountConsistency.Default,
            })) ?? [],
    };
}

export function mapFormValuesToAppStoragePayload(
    values: AppConfigStorageFormSchemaOutput,
    server: AppStorageSettings | undefined,
): AppStorageSettings {
    return {
        mounts: values.mounts.map(item => ({
            type: item.type,
            source: item.type === EMountType.Tmpfs ? "" : item.source.trim(),
            target: item.target.trim(),
            consistency: item.consistency,
        })),
        updateVer: server?.updateVer ?? 0,
    };
}
