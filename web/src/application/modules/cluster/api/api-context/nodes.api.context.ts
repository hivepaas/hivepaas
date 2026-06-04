import { createContext } from "react";

import {
    ClusterNetworksApi,
    ClusterNetworksApiValidator,
    ClusterVolumesApi,
    ClusterVolumesApiValidator,
    NodesApi,
    NodesApiValidator,
} from "~/cluster/api/services";

function createApi() {
    /**
     * Nodes
     */
    const nodesApiValidator = new NodesApiValidator();
    const clusterNetworksApiValidator = new ClusterNetworksApiValidator();
    const clusterVolumesApiValidator = new ClusterVolumesApiValidator();

    return {
        nodes: {
            $: new NodesApi(nodesApiValidator),
        },
        networks: {
            $: new ClusterNetworksApi(clusterNetworksApiValidator),
        },
        volumes: {
            $: new ClusterVolumesApi(clusterVolumesApiValidator),
        },
    };
}

export const NodesApiContext = createContext({
    api: createApi(),
});
