import { VolumeManagementTable } from "~/cluster/module-shared/components";

export function VolumesTable() {
    return <VolumeManagementTable scope={{ type: "cluster" }} />;
}
