import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ViewNetworkRoute } from "~/cluster/module-shared/components";

export function ProjectNetworkDetailsRoute() {
    const { id: projectId, networkId } = useParams<{ id: string; networkId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(networkId, "networkId must be defined");

    return (
        <ViewNetworkRoute
            scope={{ type: "project", projectId }}
            networkId={networkId}
        />
    );
}
