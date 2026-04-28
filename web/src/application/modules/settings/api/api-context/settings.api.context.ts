import { createContext } from "react";

import {
    BasicAuthApi,
    BasicAuthApiValidator,
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
    StorageSettingsApi,
    StorageSettingsApiValidator,
} from "../services";

function createApi() {
    /**
     * Settings
     */
    const notificationsValidator = new NotificationsApiValidator();
    const basicAuthValidator = new BasicAuthApiValidator();
    const domainSettingsValidator = new DomainSettingsApiValidator();
    const storageSettingsValidator = new StorageSettingsApiValidator();
    const gitCredentialsValidator = new GitCredentialsApiValidator();
    const registryAuthValidator = new RegistryAuthApiValidator();
    const sslCertValidator = new SslCertApiValidator();

    return {
        settings: {
            notifications: new NotificationsApi(notificationsValidator),
            basicAuth: new BasicAuthApi(basicAuthValidator),
            domainSettings: new DomainSettingsApi(domainSettingsValidator),
            storageSettings: new StorageSettingsApi(storageSettingsValidator),
            gitCredentials: new GitCredentialsApi(gitCredentialsValidator),
            registryAuth: new RegistryAuthApi(registryAuthValidator),
            sslCert: new SslCertApi(sslCertValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
