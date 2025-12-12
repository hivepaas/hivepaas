import { createContext } from "react";

import { NodesApi, NodesApiValidator } from "~/cluster/api/services";

function createApi() {
    /**
     * Nodes
     */
    const nodesApiValidator = new NodesApiValidator();

    return {
        nodes: {
            $: new NodesApi(nodesApiValidator),
        },
    };
}

export const NodesApiContext = createContext({
    api: createApi(),
});
