import { createContext } from "react";

import { SessionApi, SessionApiValidator } from "@application/shared/api/services";

function createApplicationApi() {
    return {
        // profile: new ProfileApi(new ProfileApiValidator()),
        session: new SessionApi(new SessionApiValidator()),
    };
}

export const ApplicationApiContext = createContext({
    api: createApplicationApi(),
});
