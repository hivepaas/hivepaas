import {
    Container,
    FileStack,
    HardDrive,
    Image,
    type LucideIcon,
    Network,
    Server,
    Terminal,
    Waypoints,
    Workflow,
} from "lucide-react";

/**
 * What the Docker API proxy lets an app do, as pkg/dockerproxy/routes.go and its
 * handlers enforce it. Written down here rather than fetched: it changes when the
 * proxy does, and the two are released together.
 */

export type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "HEAD";

export interface GuideEndpoint {
    methods: HttpMethod[];
    path: string;
    note: string;
}

export interface CoreArea {
    title: string;
    icon: LucideIcon;
    items: string[];
    endpoints: GuideEndpoint[];
}

export type RiskLevel = "low" | "medium";

export interface PermissionGuide {
    /** As the setting and the proxy name it. */
    value: string;
    label: string;
    icon: LucideIcon;
    summary: string;
    allows: string[];
    endpoints: GuideEndpoint[];
    /** Said when the group has no endpoint of its own. */
    endpointsNote?: string;
    rules: string[];
    withoutIt: string;
    usedBy: string[];
    risk: { level: RiskLevel; text: string };
}

export const CORE_AREAS: CoreArea[] = [
    {
        title: "System",
        icon: Server,
        items: ["Checking the daemon is there, and which version it runs.", "The node's info, without the cluster's."],
        endpoints: [
            { methods: ["GET", "HEAD"], path: "/_ping", note: "Is the daemon there" },
            { methods: ["GET"], path: "/version", note: "Docker's version" },
            { methods: ["GET"], path: "/info", note: "Without the swarm's details" },
        ],
    },
    {
        title: "Images",
        icon: Image,
        items: [
            "Pulling an image, only when it matches one of Images.",
            "Listing and inspecting the images of the node.",
            "Importing an image from a tarball is refused.",
        ],
        endpoints: [
            { methods: ["GET"], path: "/images/json", note: "List" },
            { methods: ["GET"], path: "/images/{name}/json", note: "Inspect" },
            { methods: ["POST"], path: "/images/create", note: "Pull, matched against Images" },
        ],
    },
    {
        title: "Its own containers",
        icon: Container,
        items: [
            "Creating a container, checked field by field: what reaches past the container is refused, the limits are applied, and it is labeled as the app's.",
            "Starting, watching, stopping and removing the containers it created - and seeing only those.",
        ],
        endpoints: [
            { methods: ["GET"], path: "/containers/json", note: "Only the app's own" },
            { methods: ["POST"], path: "/containers/create", note: "Checked, limited, labeled" },
            { methods: ["GET"], path: "/containers/{id}/json|logs|stats|top", note: "Inspect, logs, usage, processes" },
            {
                methods: ["POST"],
                path: "/containers/{id}/start|stop|kill|wait|restart|resize|attach",
                note: "Run and watch",
            },
            { methods: ["DELETE"], path: "/containers/{id}", note: "Remove" },
        ],
    },
    {
        title: "Network",
        icon: Waypoints,
        items: [
            "Every container joins the app's own network, hp-dapi-<app>. Asking for bridge or host lands there too.",
            "With Env network ticked, a container may also join the env's network, to reach the env's other apps.",
            "No network at all (none) is fine; sharing another container's network is refused.",
        ],
        endpoints: [],
    },
    {
        title: "Storage",
        icon: FileStack,
        items: [
            "tmpfs mounts.",
            "A bind of a Shared directory, which becomes a mount of the app's own directory on its volume.",
            "Any other path of the node is refused.",
        ],
        endpoints: [],
    },
];

export const PERMISSION_GUIDES: PermissionGuide[] = [
    {
        value: "exec",
        label: "Exec",
        icon: Terminal,
        summary: "Run commands inside the app's own containers, as docker exec does.",
        allows: [
            "Starting a command in a running container the app created, and attaching to its input and output.",
            "Resizing the command's terminal, and reading how the command ended.",
        ],
        endpoints: [
            { methods: ["POST"], path: "/containers/{id}/exec", note: "Create an exec" },
            { methods: ["POST"], path: "/exec/{id}/start", note: "Run it, attached or not" },
            { methods: ["POST"], path: "/exec/{id}/resize", note: "Resize its terminal" },
            { methods: ["GET"], path: "/exec/{id}/json", note: "Read its state and exit code" },
        ],
        rules: [
            "Only in containers the app created: the proxy checks the owner label of the container, and of the container an exec belongs to.",
            "Fields allowed: User, Cmd, Env, WorkingDir, Tty, AttachStdin, AttachStdout, AttachStderr, DetachKeys, ConsoleSize.",
            "Privileged is refused: an exec gets no more than its container has.",
            "The command may run as root inside the container, which itself has no added capability, no device and no path of the node.",
        ],
        withoutIt:
            "The app can start containers and read their logs, but cannot run anything in one once it has started. Every exec call is refused with 403.",
        usedBy: ["Gitea / Forgejo runner (act): every step of a job", "Jenkins Docker agents", "Coder workspaces"],
        risk: {
            level: "low",
            text: "The command runs inside a container the app already controls. Nothing reaches the node.",
        },
    },
    {
        value: "files",
        label: "Files",
        icon: FileStack,
        summary: "Copy files into and out of the app's own containers, as docker cp does.",
        allows: [
            "Uploading a tar archive into a path of a container.",
            "Downloading a path of a container as a tar archive.",
            "Checking whether a path exists, and reading its size and mode.",
        ],
        endpoints: [
            { methods: ["PUT"], path: "/containers/{id}/archive", note: "Upload into the container" },
            { methods: ["GET"], path: "/containers/{id}/archive", note: "Download from the container" },
            { methods: ["HEAD"], path: "/containers/{id}/archive", note: "Stat a path" },
        ],
        rules: [
            "Only containers the app created.",
            "What is written lands in the container's own filesystem, or in what it mounts - at most a shared directory of the app, never a path of the node.",
        ],
        withoutIt: "Files reach a container only through its image, a shared directory or a volume.",
        usedBy: [
            "Gitea / Forgejo runner (act): copies the workspace and the actions into the job's container, and results out",
        ],
        risk: {
            level: "low",
            text: "Only the containers the app created, and only what they can see.",
        },
    },
    {
        value: "volumes",
        label: "Volumes",
        icon: HardDrive,
        summary: "Keep volumes of the app's own, and mount them into its containers by name.",
        allows: [
            "Creating a volume, listing the app's volumes, inspecting and removing them.",
            "Naming a volume in a container's Binds or Mounts. One that does not exist yet is created for the app on the spot, since act names its volumes without creating them.",
        ],
        endpoints: [
            { methods: ["GET"], path: "/volumes", note: "Only the app's own" },
            { methods: ["POST"], path: "/volumes/create", note: "Local driver, labeled as the app's" },
            { methods: ["GET"], path: "/volumes/{name}", note: "Inspect one of its own" },
            { methods: ["DELETE"], path: "/volumes/{name}", note: "Remove one of its own" },
        ],
        rules: [
            "Only the local driver, on the node the app runs on, with no driver options.",
            "Every volume is labeled as the app's. Listing shows only those; inspecting, removing or mounting any other is refused - another app's data and HivePaaS's own volumes among them.",
            "A subpath of a volume must stay inside it.",
            "Volumes are not counted by the limits. They stay while the app has access, as caches, and are removed when access is turned off or the app is deleted.",
        ],
        withoutIt: "Containers can use tmpfs and the shared directories only. Naming any volume in a mount is refused.",
        usedBy: ["Gitea / Forgejo runner (act): act-toolcache, and a workspace per job"],
        risk: {
            level: "medium",
            text: "Disk space on the node: the volumes are caches that grow, and no limit bounds them.",
        },
    },
    {
        value: "networks",
        label: "Networks",
        icon: Network,
        summary: "Keep networks of the app's own, and connect its containers to them.",
        allows: [
            "Creating a network, listing, inspecting and removing it.",
            "Connecting a container to a network, or disconnecting it, with aliases and DNS names so that containers find each other by name.",
        ],
        endpoints: [
            { methods: ["GET"], path: "/networks", note: "The app's own, and those it may join" },
            { methods: ["POST"], path: "/networks/create", note: "A bridge on this node" },
            { methods: ["GET"], path: "/networks/{id}", note: "Inspect one it may use" },
            { methods: ["DELETE"], path: "/networks/{id}", note: "Remove one it created" },
            { methods: ["POST"], path: "/networks/{id}/connect", note: "Join a container" },
            { methods: ["POST"], path: "/networks/{id}/disconnect", note: "Leave" },
        ],
        rules: [
            "Only a bridge on this node: driver bridge, scope local, the default IPAM. No overlay across the cluster, no driver options, no address range of the operator's.",
            "Internal, Attachable, IPv4 and IPv6 may be set.",
            "Connecting: only the app's own containers - or the app's own task, joining its children's network - and only to the app's network, the env network when allowed, or a network the app created.",
            "An endpoint takes Aliases and DNSNames only: no static address.",
            "A network no container uses, older than an hour, is removed by the node's agent.",
        ],
        withoutIt:
            "Every container joins the app's own network - and the env network when Env network is ticked - and nothing else.",
        usedBy: [
            "Gitea / Forgejo runner (act): a network per job, so that the job's service containers (redis, postgres…) answer by name",
        ],
        risk: {
            level: "low",
            text: "The networks are local bridges of the app's own, and idle ones are removed.",
        },
    },
    {
        value: "nestedSocket",
        label: "Nested socket",
        icon: Workflow,
        summary: "Give the app's containers the same Docker API, for jobs that run docker themselves.",
        allows: [
            "Mounting the app's socket into a container: a bind of /var/run/hivepaas/docker.sock, or a mount of the socket volume.",
            "What that container starts goes through the same proxy, under the same policy.",
        ],
        endpoints: [],
        endpointsNote:
            "No endpoint of its own. It lets container create mount the app's socket: a bind of /var/run/hivepaas/docker.sock, or the volume hp-dapi-sock-<app>.",
        rules: [
            "The same images, limits and refusals at every level: --privileged, -v /:/host and the rest are refused inside as outside.",
            "Containers started from inside count toward the app's container limit.",
            "It is not Docker-in-Docker: no second daemon, nothing privileged.",
        ],
        withoutIt: "Mounting the socket into a container is refused. A job that runs docker fails.",
        usedBy: ["Gitea / Forgejo runner (act): jobs that call docker run or docker compose"],
        risk: {
            level: "medium",
            text: "The code a job runs - a pull request from anyone who can open one - can start containers of any image the app may run. With Images set to *, that is any image, within the limits.",
        },
    },
];

/** What no permission opens, whatever is ticked. */
export const NEVER_ALLOWED: string[] = [
    "--privileged",
    "Added capabilities",
    "Devices and GPUs",
    "The node's network, PID or IPC namespace",
    "Security options and runtimes",
    "Sysctls",
    "Publishing ports",
    "Any path of the node beyond the shared directories",
    "volumes-from and links",
    "docker build",
    "Swarm services and stacks",
    "Events, prune and other calls across the whole daemon",
    "Another app's containers, volumes or networks",
];

/** How the shipped templates use the permissions. */
export const TEMPLATE_EXAMPLES: { template: string; allow: string[]; why: string }[] = [
    {
        template: "Autobase",
        allow: [],
        why: "Starts one Ansible container per cluster operation, and reads its log from a shared directory. The core is enough.",
    },
    {
        template: "Gitea runner",
        allow: ["exec", "files", "volumes", "networks", "nestedSocket"],
        why: "act runs each job in containers of its own: steps by exec, the workspace by copy, caches in volumes, services on a network per job, and docker inside the job.",
    },
];
