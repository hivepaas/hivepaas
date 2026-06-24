import { useParams } from "react-router";
import { ViewNetworkRoute } from "~/cluster/module-shared/components";

export function NetworkDetailsRoute() {
    const { networkId = "" } = useParams();

    return (
        <ViewNetworkRoute
            scope={{ type: "cluster" }}
            networkId={networkId}
        />
    );
}
