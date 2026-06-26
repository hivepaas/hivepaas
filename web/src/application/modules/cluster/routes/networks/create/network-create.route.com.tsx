import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { CreateNetworkFormRoute } from "~/cluster/module-shared/components";

export function NetworkCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <CreateNetworkFormRoute scope={{ type: "cluster" }} />
        </div>
    );
}
