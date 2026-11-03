export const CAPABILITY_IDS = {
    SecretReveal: "cap::secret::reveal",
    ApiKeyCreate: "cap::api-key::create",
} as const;

export type CapabilityId = (typeof CAPABILITY_IDS)[keyof typeof CAPABILITY_IDS];

export const CAPABILITIES = [
    {
        name: "Can Reveal Secrets",
        id: CAPABILITY_IDS.SecretReveal,
    },
    {
        name: "Can Create API Key",
        id: CAPABILITY_IDS.ApiKeyCreate,
    },
] as const;
