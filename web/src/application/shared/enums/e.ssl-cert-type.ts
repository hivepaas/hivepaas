export const ESslCertType = {
    LetsEncrypt: "letsencrypt",
    Custom: "custom",
    SelfSigned: "self-signed",
} as const;

export type ESslCertType = (typeof ESslCertType)[keyof typeof ESslCertType];
