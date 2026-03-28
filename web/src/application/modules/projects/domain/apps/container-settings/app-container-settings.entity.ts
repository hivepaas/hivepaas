import type {
    EAppArmorMode,
    EHealthcheckMode,
    ERestartPolicyCondition,
    ESeccompMode,
} from "~/projects/module-shared/enums";

export type AppContainerSettings = ContainerSpec & {
    updateVer: number;
};

export type ContainerSpec = {
    labels: Record<string, string>;
    image: string;
    command: string;
    workingDir: string;
    hostname: string;
    user: string;
    groups: string[];
    stopSignal: string;
    tty: boolean;
    openStdin: boolean;
    readOnly: boolean;
    stopGracePeriod: string | null;
    privileges: Privileges | null;
    healthcheck: Healthcheck | null;
    restartPolicy: RestartPolicy | null;
};

export type RestartPolicy = {
    condition?: ERestartPolicyCondition;
    delay?: string | null;
    maxAttempts?: number | null;
    window?: string | null;
};

export type Healthcheck = {
    enabled?: boolean;
    mode?: EHealthcheckMode;
    command?: string;
    interval?: string;
    timeout?: string;
    startPeriod?: string;
    startInterval?: string;
    retries?: number;
};

export type SELinuxContext = {
    disable?: boolean;
    user?: string;
    role?: string;
    type?: string;
    level?: string;
};

export type SeccompOpts = {
    mode?: ESeccompMode;
    profile?: string;
};

export type AppArmorOpts = {
    mode?: EAppArmorMode;
};

export type Privileges = {
    seLinuxContext?: SELinuxContext | null;
    seccomp?: SeccompOpts | null;
    appArmor?: AppArmorOpts | null;
    noNewPrivileges?: boolean;
};
