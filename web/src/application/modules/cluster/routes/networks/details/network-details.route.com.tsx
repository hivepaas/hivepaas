import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { ViewNetworkRoute } from "~/cluster/module-shared/components";

export function NetworkDetailsRoute() {
    const { networkId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <ViewNetworkRoute
                scope={{ type: "cluster" }}
                networkId={networkId}
            />
        </div>
    );
}
