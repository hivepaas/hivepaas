import { createContext } from "react";

import { HomeAttentionApi, HomeAttentionApiValidator } from "../services";

function createApi() {
    return {
        home: {
            attention: new HomeAttentionApi(new HomeAttentionApiValidator()),
        },
    };
}

export const HomeApiContext = createContext({
    api: createApi(),
});
