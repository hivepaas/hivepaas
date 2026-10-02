import type { SettingsBaseEntity } from "~/settings/domain";

/** How a registry credential signs in. */
export const ERegistryAuthKind = {
    /** A username and a password, used as they are. */
    Basic: "",
    /** AWS keys, from which HivePaaS gets Amazon ECR tokens and renews them. */
    AwsEcr: "aws-ecr",
} as const;

export type ERegistryAuthKind = (typeof ERegistryAuthKind)[keyof typeof ERegistryAuthKind];

/** An Amazon ECR credential's AWS side. The secret access key comes masked. */
export interface SettingRegistryAuthEcr {
    region: string;
    accessKeyId: string;
    secretAccessKey: string;
    roleArn: string;
    /** When the token kept for the credential expires; null before one is got. */
    tokenExpiresAt: Date | null;
}

export interface SettingRegistryAuth extends SettingsBaseEntity {
    kind: string;
    address: string;
    username: string;
    password: string;
    readonly: boolean;
    secretMasked?: boolean;
    ecr: SettingRegistryAuthEcr | null;
}
