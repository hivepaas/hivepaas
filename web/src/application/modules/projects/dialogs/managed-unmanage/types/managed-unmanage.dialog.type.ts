/** The managed database to make an ordinary app. */
export interface ManagedUnmanageTarget {
    projectId: string;
    env: string;
    appId: string;
    appName: string;
    /** Its engine, which the copy names what it archives after. */
    engine: string;
    /** Its restore test runs in a verifier app, which goes with the engine. */
    hasVerifier: boolean;
}

export interface ManagedUnmanageDialogState {
    state: { mode: "open"; target: ManagedUnmanageTarget } | { mode: "closed"; target: null };
}

export interface ManagedUnmanageDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
