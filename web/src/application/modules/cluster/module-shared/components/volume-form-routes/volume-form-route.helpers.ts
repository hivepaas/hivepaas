import type { ClusterVolume, ClusterVolumeBasePayload, ClusterVolumeCreatePayload } from "~/cluster/domain";
import {
    EClusterVolumeDriverMode,
    EClusterVolumeLocalType,
    EClusterVolumePropagation,
} from "~/cluster/module-shared/enums";
import { getClusterVolumeLocalType, toKeyValueList, toRecord } from "~/cluster/module-shared/utils";

import {
    type CreateOrEditVolumeFormInput,
    type CreateOrEditVolumeFormOutput,
    DEFAULT_VOLUME_FORM_VALUES,
} from "../volume-form";

function optionalString(value: string) {
    const trimmed = value.trim();

    return trimmed || undefined;
}

function optionalInteger(value: string) {
    const trimmed = value.trim();
    if (!trimmed) {
        return undefined;
    }

    const parsed = Number(trimmed);

    return Number.isFinite(parsed) ? parsed : undefined;
}

export function toVolumeFormInitialValues(volume: ClusterVolume): CreateOrEditVolumeFormInput {
    const isLocalDriver = volume.driver === "local";

    return {
        ...DEFAULT_VOLUME_FORM_VALUES,
        name: volume.name,
        driverMode: isLocalDriver ? EClusterVolumeDriverMode.Local : EClusterVolumeDriverMode.Custom,
        customDriverName: isLocalDriver ? "" : volume.driver,
        localType: getClusterVolumeLocalType(volume) ?? EClusterVolumeLocalType.Bind,
        bindOptions: {
            ...DEFAULT_VOLUME_FORM_VALUES.bindOptions,
            directory: volume.bindOptions?.directory ?? DEFAULT_VOLUME_FORM_VALUES.bindOptions.directory,
            propagation: volume.bindOptions?.propagation ?? EClusterVolumePropagation.Default,
            readonly: volume.bindOptions?.readonly ?? false,
            extraOptions: volume.bindOptions?.extraOptions ?? DEFAULT_VOLUME_FORM_VALUES.bindOptions.extraOptions,
        },
        nfsOptions: {
            ...DEFAULT_VOLUME_FORM_VALUES.nfsOptions,
            addr: volume.nfsOptions?.addr ?? "",
            device: volume.nfsOptions?.device ?? "",
            version: volume.nfsOptions?.version ?? DEFAULT_VOLUME_FORM_VALUES.nfsOptions.version,
            readonly: volume.nfsOptions?.readonly ?? false,
            extraOptions: volume.nfsOptions?.extraOptions ?? DEFAULT_VOLUME_FORM_VALUES.nfsOptions.extraOptions,
        },
        tmpfsOptions: {
            ...DEFAULT_VOLUME_FORM_VALUES.tmpfsOptions,
            device: volume.tmpfsOptions?.device ?? "",
            size: volume.tmpfsOptions?.size ?? DEFAULT_VOLUME_FORM_VALUES.tmpfsOptions.size,
            uid:
                volume.tmpfsOptions?.uid === undefined
                    ? DEFAULT_VOLUME_FORM_VALUES.tmpfsOptions.uid
                    : String(volume.tmpfsOptions.uid),
            extraOptions: volume.tmpfsOptions?.extraOptions ?? "",
        },
        labels: toKeyValueList(volume.labels),
        options: toKeyValueList(volume.options),
        availableInProjects: volume.availableInProjects ?? false,
        default: volume.default ?? false,
    };
}

export function toVolumeBasePayload(values: CreateOrEditVolumeFormOutput): ClusterVolumeBasePayload {
    const driver =
        values.driverMode === EClusterVolumeDriverMode.Local ? EClusterVolumeDriverMode.Local : values.customDriverName;
    const payload: ClusterVolumeBasePayload = {
        name: values.name,
        driver,
        labels: toRecord(values.labels),
        options: toRecord(values.options),
    };

    if (values.driverMode !== EClusterVolumeDriverMode.Local) {
        return payload;
    }

    if (values.localType === EClusterVolumeLocalType.Bind) {
        return {
            ...payload,
            bindOptions: {
                directory: optionalString(values.bindOptions.directory),
                propagation: values.bindOptions.propagation,
                readonly: values.bindOptions.readonly,
                extraOptions: optionalString(values.bindOptions.extraOptions),
            },
        };
    }

    if (values.localType === EClusterVolumeLocalType.Nfs) {
        return {
            ...payload,
            nfsOptions: {
                addr: values.nfsOptions.addr,
                device: values.nfsOptions.device,
                version: optionalString(values.nfsOptions.version),
                readonly: values.nfsOptions.readonly,
                extraOptions: optionalString(values.nfsOptions.extraOptions),
            },
        };
    }

    return {
        ...payload,
        tmpfsOptions: {
            device: optionalString(values.tmpfsOptions.device),
            size: values.tmpfsOptions.size,
            uid: optionalInteger(values.tmpfsOptions.uid),
            extraOptions: optionalString(values.tmpfsOptions.extraOptions),
        },
    };
}

export function toVolumeCreatePayload(
    values: CreateOrEditVolumeFormOutput,
    availableInProjects: boolean,
): ClusterVolumeCreatePayload {
    return {
        ...toVolumeBasePayload(values),
        availableInProjects,
        default: values.default,
    };
}
