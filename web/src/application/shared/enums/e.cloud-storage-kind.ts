export const ECloudStorageKind = {
    AWSS3: "aws-s3",
} as const;

export type ECloudStorageKind = (typeof ECloudStorageKind)[keyof typeof ECloudStorageKind];
