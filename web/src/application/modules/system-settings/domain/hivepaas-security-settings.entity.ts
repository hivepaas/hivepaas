export type HivePaaSSecuritySettings = {
    returnSecretsViaApi: boolean;
};

export type HivePaaSSecuritySettingsUpdatePayload = {
    appSecret: string;
    returnSecretsViaApi: boolean;
};
