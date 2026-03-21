import { createContext } from "react";

import {
    GitCredentialsApi,
    GitCredentialsApiValidator,
    NotificationsApi,
    NotificationsApiValidator,
    RegistryAuthApi,
    RegistryAuthApiValidator,
} from "../services";

function createApi() {
    /**
     * Settings
     */
    const notificationsValidator = new NotificationsApiValidator();
    const gitCredentialsValidator = new GitCredentialsApiValidator();
    const registryAuthValidator = new RegistryAuthApiValidator();

    return {
        settings: {
            notifications: new NotificationsApi(notificationsValidator),
            gitCredentials: new GitCredentialsApi(gitCredentialsValidator),
            registryAuth: new RegistryAuthApi(registryAuthValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
