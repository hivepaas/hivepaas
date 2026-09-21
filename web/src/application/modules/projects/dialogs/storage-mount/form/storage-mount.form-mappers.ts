import type { AppStorageMount } from "~/projects/domain";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import type { StorageMountFormInput, StorageMountFormOutput } from "../schemas";

export function mountToFormInput(mount: AppStorageMount): StorageMountFormInput {
    const volumeOpts = mount.volumeOptions ?? mount.clusterOptions;

    return {
        source: mount.source ?? "",
        subpath: volumeOpts?.subpath ?? "",
        readOnly: mount.readOnly ?? false,
        noCopy: volumeOpts?.noCopy ?? false,
        target: mount.target ?? "",
        consistency: mount.consistency ?? EMountConsistency.Default,
        sourceAppId: mount.sourceApp?.appId ?? "",
        sourceAppWrite: mount.sourceApp?.write ?? false,
    };
}

export function formValuesToMount(values: StorageMountFormOutput): AppStorageMount {
    const sourceApp = values.sourceAppId
        ? { appId: values.sourceAppId, write: values.sourceAppWrite ?? false }
        : undefined;

    return {
        type: EMountType.Volume,
        source: values.source,
        target: values.target,
        // For another app's directory the answer is the one stated beside the app,
        // which is what the API reads; sending both would be two answers.
        readOnly: sourceApp ? !sourceApp.write : values.readOnly,
        consistency: values.consistency,
        volumeOptions: {
            subpath: values.subpath ?? "",
            noCopy: values.noCopy ?? false,
        },
        sourceApp,
    };
}

export const emptyStorageMountFormDefaults = {
    source: "",
    subpath: "",
    readOnly: false,
    noCopy: false,
    target: "",
    consistency: EMountConsistency.Default,
    sourceAppId: "",
    sourceAppWrite: false,
};
