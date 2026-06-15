export const ESslProviderKind = {
    LetsEncrypt: "letsencrypt",
    ZeroSSL: "zerossl",
    GoogleTrust: "googletrust",
} as const;

export type ESslProviderKind = (typeof ESslProviderKind)[keyof typeof ESslProviderKind];
