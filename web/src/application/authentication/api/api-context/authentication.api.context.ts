import { createContext } from "react";

import { AuthApi, AuthApiMapper, AuthApiValidator } from "@application/authentication/api/services";

function createApi() {
    return {
        auth: new AuthApi(new AuthApiValidator(), new AuthApiMapper()),
    };
}

export const AuthenticationApiContext = createContext({
    api: createApi(),
});
