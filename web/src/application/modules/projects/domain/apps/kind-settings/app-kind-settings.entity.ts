export type EAppCategory = "webapp" | "database" | "cache" | "storage";

export const ALL_APP_CATEGORIES: EAppCategory[] = ["webapp", "database", "cache", "storage"];

export type EDatabaseSSLMode = "disable" | "prefer" | "require" | "verify-ca" | "verify-full";

export const ALL_DATABASE_SSL_MODES: EDatabaseSSLMode[] = ["disable", "prefer", "require", "verify-ca", "verify-full"];

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface AppKindWebapp {}

export interface AppKindDatabase {
    dbName: string;
    username: string;
    password?: string;
    rootPassword?: string;
    sslMode: EDatabaseSSLMode;
    sslCert?: { id: string; name?: string } | null;
    tlsPassthrough?: boolean;
}

export interface AppKindCache {
    password?: string;
    maxMemory?: string;
    evictionRule?: string;
    persistenceMode?: string;
    sslCert?: { id: string; name?: string } | null;
}

export interface AppKindStorage {
    keyId?: string;
    secret?: string;
    bucket?: string;
    region?: string;
}

export interface AppKindSettings {
    category: EAppCategory;
    engine: string;
    port: number;
    version: string;
    webapp?: AppKindWebapp | null;
    database?: AppKindDatabase | null;
    cache?: AppKindCache | null;
    storage?: AppKindStorage | null;
    secretMasked: boolean;
    updateVer: number;
}

export interface AppKindSettingsUpdatePayload {
    category: EAppCategory;
    engine: string;
    port: number;
    version: string;
    webapp?: AppKindWebapp;
    database?: {
        dbName: string;
        username: string;
        password?: string;
        rootPassword?: string;
        sslMode: EDatabaseSSLMode;
        sslCert?: { id: string } | null;
        tlsPassthrough?: boolean;
    };
    cache?: {
        password?: string;
        maxMemory?: string;
        evictionRule?: string;
        persistenceMode?: string;
        sslCert?: { id: string } | null;
    };
    storage?: {
        keyId?: string;
        secret?: string;
        bucket?: string;
        region?: string;
    };
    updateVer: number;
}
