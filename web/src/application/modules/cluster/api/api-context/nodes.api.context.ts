import { createContext } from "react";

import { ClusterVolumesApi, ClusterVolumesApiValidator, NodesApi, NodesApiValidator } from "~/cluster/api/services";

function createApi() {
    /**
     * Nodes
     */
    const nodesApiValidator = new NodesApiValidator();
    const clusterVolumesApiValidator = new ClusterVolumesApiValidator();

    return {
        nodes: {
            $: new NodesApi(nodesApiValidator),
        },
        volumes: {
            $: new ClusterVolumesApi(clusterVolumesApiValidator),
        },
    };
}

export const NodesApiContext = createContext({
    api: createApi(),
});
