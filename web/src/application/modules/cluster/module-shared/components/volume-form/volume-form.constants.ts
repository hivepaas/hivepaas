import {
    EClusterVolumeDriverMode,
    EClusterVolumeLocalType,
    EClusterVolumePropagation,
} from "~/cluster/module-shared/enums";

import type { CreateOrEditVolumeFormInput } from "./create-or-edit-volume.form.schema";

/**
 * What the server resolves to the node it is running on.
 *
 * Sent instead of a concrete id so the answer is worked out server-side, where
 * "the current node" is actually known.
 */
export const CURRENT_NODE_VALUE = "current";

export const DEFAULT_VOLUME_FORM_VALUES: CreateOrEditVolumeFormInput = {
    name: "",
    driverMode: EClusterVolumeDriverMode.Local,
    customDriverName: "",
    // Pinned to the current node unless the operator says otherwise: a volume
    // that turns out to live on one node and claims to live everywhere is the
    // failure that only shows up later, when a task cannot find its data.
    nodeId: CURRENT_NODE_VALUE,
    nodeLabel: "",
    localType: EClusterVolumeLocalType.Bind,
    bindOptions: {
        directory: "",
        propagation: EClusterVolumePropagation.Default,
        readonly: false,
        extraOptions: "",
    },
    nfsOptions: {
        addr: "",
        device: "",
        version: "",
        readonly: false,
        extraOptions: "",
    },
    tmpfsOptions: {
        device: "",
        size: "",
        uid: "",
        extraOptions: "",
    },
    labels: [],
    options: [],
    inheritable: false,
    default: false,
};
