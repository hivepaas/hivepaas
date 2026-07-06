import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { CreateVolumeFormRoute } from "~/cluster/module-shared/components";

export function VolumeCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <CreateVolumeFormRoute scope={{ type: "cluster" }} />
        </div>
    );
}
