export const EUserRole = {
    Owner: "owner",
    Admin: "admin",
    Member: "member",
} as const;

export type EUserRole = (typeof EUserRole)[keyof typeof EUserRole];
