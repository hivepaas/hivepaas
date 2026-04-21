import { createContext } from "react";

import {
    DomainSettingsApi,
    DomainSettingsApiValidator,
    GitCredentialsApi,
    GitCredentialsApiValidator,
    NotificationsApi,
    NotificationsApiValidator,
    RegistryAuthApi,
    RegistryAuthApiValidator,
    SslCertApi,
    SslCertApiValidator,
} from "../services";

function createApi() {
    /**
     * Settings
     */
    const notificationsValidator = new NotificationsApiValidator();
    const domainSettingsValidator = new DomainSettingsApiValidator();
    const gitCredentialsValidator = new GitCredentialsApiValidator();
    const registryAuthValidator = new RegistryAuthApiValidator();
    const sslCertValidator = new SslCertApiValidator();

    return {
        settings: {
            notifications: new NotificationsApi(notificationsValidator),
            domainSettings: new DomainSettingsApi(domainSettingsValidator),
            gitCredentials: new GitCredentialsApi(gitCredentialsValidator),
            registryAuth: new RegistryAuthApi(registryAuthValidator),
            sslCert: new SslCertApi(sslCertValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
