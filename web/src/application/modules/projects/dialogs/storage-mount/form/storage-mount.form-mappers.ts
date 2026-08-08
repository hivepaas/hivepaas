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
    };
}

export function formValuesToMount(values: StorageMountFormOutput): AppStorageMount {
    return {
        type: EMountType.Volume,
        source: values.source,
        target: values.target,
        readOnly: values.readOnly,
        consistency: values.consistency,
        volumeOptions: {
            subpath: values.subpath ?? "",
            noCopy: values.noCopy ?? false,
        },
    };
}

export const emptyStorageMountFormDefaults = {
    source: "",
    subpath: "",
    readOnly: false,
    noCopy: false,
    target: "",
    consistency: EMountConsistency.Default,
};
