import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ViewNetworkRoute } from "~/cluster/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

export function ProjectNetworkDetailsRoute() {
    const { id: projectId, networkId } = useParams<{ id: string; networkId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(networkId, "networkId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <ViewNetworkRoute
            scope={{ type: "project", projectId, env: scopedEnv }}
            networkId={networkId}
        />
    );
}
