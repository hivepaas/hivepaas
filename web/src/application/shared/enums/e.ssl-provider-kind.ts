export const ESslProviderKind = {
    LetsEncrypt: "letsencrypt",
    ZeroSSL: "zerossl",
    GoogleTrustServices: "googlets",
} as const;

export type ESslProviderKind = (typeof ESslProviderKind)[keyof typeof ESslProviderKind];
