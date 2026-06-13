import { ESslKeyType, ESslProviderKind } from "@application/shared/enums";

export const SSL_PROVIDER_OPTIONS = [
    {
        value: ESslProviderKind.LetsEncrypt,
        label: "Let’s Encrypt",
    },
    {
        value: ESslProviderKind.ZeroSSL,
        label: "Zero SSL",
    },
    {
        value: ESslProviderKind.GoogleTrustServices,
        label: "Google Trust Services",
    },
] as const;

export const SSL_KEY_TYPE_OPTIONS = [
    {
        value: ESslKeyType.ECP256,
        label: "ECDSA P256 (ec-p256)",
    },
    {
        value: ESslKeyType.ECP384,
        label: "ECDSA P384 (ec-p384)",
    },
    {
        value: ESslKeyType.ECP521,
        label: "ECDSA P521 (ec-p521)",
    },
    {
        value: ESslKeyType.RSA2048,
        label: "RSA 2048 (rsa-2048)",
    },
    {
        value: ESslKeyType.RSA3072,
        label: "RSA 3072 (rsa-3072)",
    },
    {
        value: ESslKeyType.RSA4096,
        label: "RSA 4096 (rsa-4096)",
    },
    {
        value: ESslKeyType.RSA8192,
        label: "RSA 8192 (rsa-8192)",
    },
] as const;

export function formatSslProviderKind(kind: string): string {
    return SSL_PROVIDER_OPTIONS.find(option => option.value === kind)?.label ?? kind;
}
