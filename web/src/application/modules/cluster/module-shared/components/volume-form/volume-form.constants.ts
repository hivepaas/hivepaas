import {
    EClusterVolumeDriverMode,
    EClusterVolumeLocalType,
    EClusterVolumePropagation,
} from "~/cluster/module-shared/enums";

import type { CreateOrEditVolumeFormInput } from "./create-or-edit-volume.form.schema";

export const DEFAULT_VOLUME_FORM_VALUES: CreateOrEditVolumeFormInput = {
    name: "",
    driverMode: EClusterVolumeDriverMode.Local,
    customDriverName: "",
    localType: EClusterVolumeLocalType.Bind,
    bindOptions: {
        directory: "auto",
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
    availableInProjects: false,
    default: false,
};
