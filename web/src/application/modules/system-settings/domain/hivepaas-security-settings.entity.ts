/** An app given the node's own Docker socket, with where to find it. */
export type PrivilegedApp = {
    appId: string;
    appName: string;
    projectId: string;
    projectName: string;
    projectEnvKey: string;
    projectEnvName: string;
};

export type HivePaaSSecuritySettings = {
    returnSecretsViaApi: boolean;
    /** Kinds of secret returned whatever returnSecretsViaApi says. */
    alwaysReturnSecretTypes: string[];
    /**
     * Whether an administrator may give an app raw access to the host it runs on:
     * the node's Docker socket, or a directory of the node through an import.
     */
    allowPrivilegedApps: boolean;
    /** The apps given the node's Docker socket. Turning the switch off takes it from none of them. */
    privilegedApps: PrivilegedApp[];
};

/**
 * The whole new state, not a patch: a field left out is saved as off. Every save
 * carries the current value of the switches it does not change.
 */
export type HivePaaSSecuritySettingsUpdatePayload = {
    appSecret: string;
    returnSecretsViaApi: boolean;
    alwaysReturnSecretTypes: string[];
    allowPrivilegedApps: boolean;
};
