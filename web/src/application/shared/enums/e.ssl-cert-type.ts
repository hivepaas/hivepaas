export const ESslCertType = {
    GoogleTrust: "googletrust",
    LetsEncrypt: "letsencrypt",
    ZeroSSL: "zerossl",
    Custom: "custom",
    SelfSigned: "self-signed",
} as const;

export type ESslCertType = (typeof ESslCertType)[keyof typeof ESslCertType];
