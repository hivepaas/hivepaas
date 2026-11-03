import type { SettingsBaseEntity } from "~/settings/domain";

/** Where the images are kept. Chosen when the registry is provisioned and refused afterwards. */
export const REGISTRY_STORAGE_TYPES = ["volume", "s3"] as const;
export type HivePaaSRegistryStorageType = (typeof REGISTRY_STORAGE_TYPES)[number];

export type HivePaaSRegistryStorage = {
    type: HivePaaSRegistryStorageType;
    /** The cluster volume, for type "volume". It also decides which node the registry runs on. */
    volume?: { id: string } | null;
    /** The cloud storage, for type "s3". Bucket, region and credentials come from it. */
    cloudStorage?: { id: string } | null;
};

export type HivePaaSRegistryCleanup = {
    enabled: boolean;
    /** "policy" today: the two numbers below, turned into the registry's own retention rules. */
    mode: string;
    /** How many of the newest builds of every app are kept, however old they are. */
    keepLast: number;
    /** Everything pushed, or pulled, within this many days is kept. */
    keepDays: number;
};

/** What is running, as opposed to what was configured. */
export type HivePaaSRegistryStatus = {
    provisioned: boolean;
    appId?: string;
    reachable: boolean;
    /** Why it could not be reached. Normal while it restarts after a save. */
    unreachable?: string;
    repositories: number;
    storedBytes: number;
};

export type HivePaaSRegistryCredentialRotation = {
    rotatedAt?: string;
    /** When the previous password stops working: the date every app has to be redeployed by. */
    graceEndsAt?: string;
};

export interface HivePaaSRegistrySettings extends SettingsBaseEntity {
    enabled: boolean;
    type: string;
    managed: boolean;
    domain: string;
    storage: HivePaaSRegistryStorage;
    cleanup: HivePaaSRegistryCleanup;
    /** Whether zot's own web interface answers at the registry's domain. */
    dashboardEnabled: boolean;
    /** CPU cores the registry may use. Read off its service, not stored. */
    cpuLimit: number;
    /** A size carrying its unit, such as "512mb". Read off its service, not stored. */
    memoryLimit: string;
    /** The provisioned app, absent until the registry has been provisioned. */
    app?: { id: string } | null;
    /** The registry auth setting every push and pull uses. */
    credential?: { id: string } | null;
    registryStatus?: HivePaaSRegistryStatus;
    credentialRotation?: HivePaaSRegistryCredentialRotation | null;
}

export type HivePaaSRegistryDomainProbe = {
    reached: boolean;
    proxied: boolean;
    /** Sentences naming what said so, for the operator to act on. */
    evidence: string[];
};

export type HivePaaSRegistryPushCheck = {
    ok: boolean;
    statusCode: number;
    detail: string;
    elapsedMs: number;
};
