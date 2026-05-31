import { createContext } from "react";

import {
    AccessTokenApi,
    AccessTokenApiValidator,
    BasicAuthApi,
    BasicAuthApiValidator,
    CloudStorageApi,
    CloudStorageApiValidator,
    DomainSettingsApi,
    DomainSettingsApiValidator,
    EmailApi,
    EmailApiValidator,
    GitCredentialsApi,
    GitCredentialsApiValidator,
    GithubAppApi,
    GithubAppApiValidator,
    ImServiceApi,
    ImServiceApiValidator,
    NotificationsApi,
    NotificationsApiValidator,
    OAuthApi,
    OAuthApiValidator,
    RegistryAuthApi,
    RegistryAuthApiValidator,
    RepoWebhookApi,
    RepoWebhookApiValidator,
    SSHKeyApi,
    SSHKeyApiValidator,
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
    const sshKeyValidator = new SSHKeyApiValidator();
    const accessTokenValidator = new AccessTokenApiValidator();
    const cloudStorageValidator = new CloudStorageApiValidator();
    const oauthValidator = new OAuthApiValidator();
    const githubAppValidator = new GithubAppApiValidator();
    const repoWebhookValidator = new RepoWebhookApiValidator();

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
            sshKey: new SSHKeyApi(sshKeyValidator),
            accessToken: new AccessTokenApi(accessTokenValidator),
            cloudStorage: new CloudStorageApi(cloudStorageValidator),
            oauth: new OAuthApi(oauthValidator),
            githubApp: new GithubAppApi(githubAppValidator),
            repoWebhook: new RepoWebhookApi(repoWebhookValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
