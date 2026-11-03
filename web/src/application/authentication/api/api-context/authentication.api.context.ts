import { createContext } from "react";

import { DeviceInfo } from "@infrastructure/device";

import { AuthApi, AuthApiValidator } from "@application/authentication/api/services";

function createApi() {
    return {
        auth: new AuthApi(new AuthApiValidator(), new DeviceInfo()),
    };
}

export const AuthenticationApiContext = createContext({
    api: createApi(),
});
