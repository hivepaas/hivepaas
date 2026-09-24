export type HivePaaSSecuritySettings = {
    returnSecretsViaApi: boolean;
    /** Kinds of secret returned whatever returnSecretsViaApi says. */
    alwaysReturnSecretTypes: string[];
    /** Whether apps may reach into the host they run on: the Docker socket, a directory of the node. */
    allowPrivilegedApps: boolean;
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
