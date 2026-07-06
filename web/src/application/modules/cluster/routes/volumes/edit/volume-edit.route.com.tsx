import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { EditVolumeFormRoute } from "~/cluster/module-shared/components";

export function VolumeEditRoute() {
    const { volumeId } = useParams<{ volumeId: string }>();

    invariant(volumeId, "volumeId must be defined");

    return (
        <div className={cn(listBox)}>
            <EditVolumeFormRoute
                scope={{ type: "cluster" }}
                volumeId={volumeId}
            />
        </div>
    );
}
