export const AppDataFileStorageType = {
    Local: "local",
    Cloud: "cloud",
} as const;

export type AppDataFileStorageType = (typeof AppDataFileStorageType)[keyof typeof AppDataFileStorageType];

export interface AppDataFileStorage {
    name: string;
}

export interface AppDataFile {
    id: string;
    type: string;
    kind: string;
    key: string;
    status: string;
    name: string;
    path: string;
    bucket?: string;
    mimetype: string;
    sizeBytes: number;
    storageType: AppDataFileStorageType;
    storage: AppDataFileStorage | null;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date;
}
