export const ESSHKeyType = {
    Ed25519: "ed25519",
    ECP256: "ec-p256",
    ECP384: "ec-p384",
    ECP521: "ec-p521",
    RSA2048: "rsa-2048",
    RSA3072: "rsa-3072",
    RSA4096: "rsa-4096",
    RSA8192: "rsa-8192",
} as const;

export type ESSHKeyType = (typeof ESSHKeyType)[keyof typeof ESSHKeyType];
