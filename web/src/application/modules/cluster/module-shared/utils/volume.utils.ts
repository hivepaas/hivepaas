import type { ClusterVolume } from "~/cluster/domain";
import { EClusterVolumeLocalType } from "~/cluster/module-shared/enums";

export function getClusterVolumeLocalType(volume: ClusterVolume): EClusterVolumeLocalType | null {
    if (volume.driver !== "local") {
        return null;
    }

    if (volume.bindOptions) {
        return EClusterVolumeLocalType.Bind;
    }

    if (volume.nfsOptions) {
        return EClusterVolumeLocalType.Nfs;
    }

    if (volume.tmpfsOptions) {
        return EClusterVolumeLocalType.Tmpfs;
    }

    return null;
}

export function getClusterVolumeTypeLabel(volume: ClusterVolume): string {
    return getClusterVolumeLocalType(volume) ?? "-";
}

export function toKeyValueList(map?: Record<string, string> | null): { key: string; value: string }[] {
    return Object.entries(map ?? {}).map(([key, value]) => ({ key, value }));
}

export function toRecord(items: { key: string; value: string }[]): Record<string, string> {
    return items.reduce<Record<string, string>>((acc, item) => {
        const key = item.key.trim();
        if (!key) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, {});
}
