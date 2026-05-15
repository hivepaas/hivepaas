import { createContext } from "react";

import {
    BasicAuthApi,
    BasicAuthApiValidator,
    DomainSettingsApi,
    DomainSettingsApiValidator,
    EmailApi,
    EmailApiValidator,
    GitCredentialsApi,
    GitCredentialsApiValidator,
    ImServiceApi,
    ImServiceApiValidator,
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
    const imServiceValidator = new ImServiceApiValidator();
    const emailValidator = new EmailApiValidator();

    return {
        settings: {
            notifications: new NotificationsApi(notificationsValidator),
            basicAuth: new BasicAuthApi(basicAuthValidator),
            domainSettings: new DomainSettingsApi(domainSettingsValidator),
            storageSettings: new StorageSettingsApi(storageSettingsValidator),
            gitCredentials: new GitCredentialsApi(gitCredentialsValidator),
            registryAuth: new RegistryAuthApi(registryAuthValidator),
            sslCert: new SslCertApi(sslCertValidator),
            imService: new ImServiceApi(imServiceValidator),
            email: new EmailApi(emailValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
