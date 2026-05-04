import type { AppStorageMount } from "~/projects/domain";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import type { StorageMountFormInput, StorageMountFormOutput } from "../schemas";

const recordToKvArray = (rec?: Record<string, string> | null) =>
    rec ? Object.entries(rec).map(([key, value]) => ({ key, value })) : [];

const kvArrayToRecord = (arr?: { key: string; value: string }[]): Record<string, string> | undefined =>
    arr && arr.length > 0
        ? Object.fromEntries(arr.filter(({ key }) => key.trim() !== "").map(({ key, value }) => [key, value]))
        : undefined;

export function mountToFormInput(mount: AppStorageMount): StorageMountFormInput {
    if (mount.type === EMountType.Volume) {
        return {
            ...mount,
            volumeOptions: {
                ...mount.volumeOptions,
                labels: recordToKvArray(mount.volumeOptions?.labels),
            },
        } as StorageMountFormInput;
    }
    if (mount.type === EMountType.Cluster) {
        return {
            ...mount,
            clusterOptions: {
                ...mount.clusterOptions,
                labels: recordToKvArray(mount.clusterOptions?.labels),
            },
        } as StorageMountFormInput;
    }
    return mount as StorageMountFormInput;
}

export function formValuesToMount(values: StorageMountFormOutput): AppStorageMount {
    if (values.type === EMountType.Volume) {
        return {
            ...values,
            volumeOptions: {
                ...values.volumeOptions,
                labels: kvArrayToRecord(values.volumeOptions.labels),
            },
        };
    }
    if (values.type === EMountType.Cluster) {
        return {
            ...values,
            clusterOptions: {
                ...values.clusterOptions,
                labels: kvArrayToRecord(values.clusterOptions.labels),
            },
        };
    }
    return values;
}

export const emptyStorageMountFormDefaults: StorageMountFormInput = {
    type: EMountType.Bind,
    target: "",
    readOnly: false,
    consistency: EMountConsistency.Default,
    bindOptions: {
        baseDir: "",
        subpath: "",
        subpathRequired: "",
    },
};
