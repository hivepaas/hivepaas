/** How an app reaches the Docker API: through the proxy HivePaaS runs, or the node's own socket. */
export type AppDockerApiMode = "proxy" | "host";

/** What keeps the person from choosing host mode: the privileged-apps switch, then being an administrator. */
export type AppDockerApiHostModeBlocker = "" | "switch" | "admin";

/** Limits of the app's children. Zero, and an empty memory, stand for the default. */
export type AppDockerApiLimits = {
    containers: number;
    /** A data size such as "2gb". */
    memory: string;
    cpus: number;
};

/**
 * What an app may do through the Docker API. Access that is turned off keeps
 * what it allowed, and host mode keeps the proxy's policy, for going back.
 */
export interface AppDockerApiSettings {
    enabled: boolean;
    mode: AppDockerApiMode;
    /** Patterns over the images children may run; "*" is any image. */
    images: string[];
    /** Directories of the app a child may bind. */
    sharedDirs: string[];
    /** Volume names a child may mount, each standing for one of sharedDirs. */
    sharedVolumes: Record<string, string>;
    /** "env": children may also join the app's env network. */
    networks: string[];
    /** Groups of endpoints beyond the core. */
    allow: string[];
    limits: AppDockerApiLimits;
    defaultLimits: AppDockerApiLimits;
    hostMode: {
        available: boolean;
        blockedBy: AppDockerApiHostModeBlocker;
    };
    updateVer: number;
}
