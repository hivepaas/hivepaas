/**
 * How many requests and connections OBI tracks at once on a node: the size of its eBPF maps, which it
 * allocates whole when it starts, and so its memory.
 */
export type LoggingPerformanceCapacity = "small" | "medium" | "large";

/** A node's capacity as chosen: `auto` for the one HivePaaS recommends for the node's memory. */
export type LoggingPerformanceCapacityChoice = LoggingPerformanceCapacity | "auto";

/** A capacity a node can be given: about how much memory OBI takes at it, and how many it tracks at once. */
export interface LoggingPerformanceCapacityInfo {
    capacity: LoggingPerformanceCapacity;
    memoryMiB: number;
    tracked: number;
}

/** What a node's agent last said of it, within the last few minutes. */
export interface LoggingPerformanceNodeStatus {
    time: string;
    /** The settings ask for OBI on the node. */
    wanted: boolean;
    running: boolean;
    /** The apps OBI watches there. */
    apps: number;
    /**
     * The node can run OBI; reasons say why not: kernel-too-old, no-btf, container-virt, no-tracefs,
     * lockdown, low-memory.
     */
    ok: boolean;
    reasons: string[];
    kernel: string;
    memTotalMb: number;
    memAvailableMb: number;
    /** The capacity OBI runs, or would run, with there. */
    capacity: string;
}

/** A node of the swarm: whether it runs OBI, at which capacity, and what its agent last said of it. */
export interface LoggingPerformanceNode {
    id: string;
    hostname: string;
    role: string;
    state: string;
    availability: string;
    memoryBytes: number;
    /** The capacity HivePaaS recommends for the node's memory. */
    recommended: LoggingPerformanceCapacity;
    enabled: boolean;
    capacity: LoggingPerformanceCapacityChoice;
    status: LoggingPerformanceNodeStatus | null;
}

/** Why the nodes' statuses were not read: the agents say nothing while the feature is off. */
export type LoggingPerformanceStatusReason = "off" | "logs-not-stored" | "unreadable";

/** The collection of apps' routes and calls by OBI (eBPF), and the nodes that run it. */
export interface LoggingPerformance {
    enabled: boolean;
    /** The logging settings were saved: these are saved with them, and cannot be before. */
    configured: boolean;
    /** OBI runs only while the logs are stored: its numbers are rows in them. */
    logsStored: boolean;
    /** The logging settings' version: saving these changes it, as saving the logging settings does. */
    updateVer: number;
    statusReason: LoggingPerformanceStatusReason | null;
    capacities: LoggingPerformanceCapacityInfo[];
    nodes: LoggingPerformanceNode[];
}
