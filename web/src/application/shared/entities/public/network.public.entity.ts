export interface NetworkPublic {
    id: string;
    name: string;
    availableInProjects: boolean;
    driver: string;
    internal: boolean;
    attachable: boolean;
    ingress: boolean;
    enableIPv4: boolean;
    enableIPv6: boolean;
    options: Record<string, string>;
    labels: Record<string, string>;
    createdAt: Date;
}
