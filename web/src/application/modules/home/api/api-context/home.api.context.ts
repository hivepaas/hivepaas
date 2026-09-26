import { createContext } from "react";

import { GetStartedApi, GetStartedApiValidator, HomeAttentionApi, HomeAttentionApiValidator } from "../services";

function createApi() {
    return {
        home: {
            attention: new HomeAttentionApi(new HomeAttentionApiValidator()),
            getStarted: new GetStartedApi(new GetStartedApiValidator()),
        },
    };
}

export const HomeApiContext = createContext({
    api: createApi(),
});
