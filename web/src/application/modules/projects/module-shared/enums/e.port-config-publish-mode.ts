export const EPortConfigPublishMode = {
    Ingress: "ingress",
    Host: "host",
} as const;

export type EPortConfigPublishMode = (typeof EPortConfigPublishMode)[keyof typeof EPortConfigPublishMode];
