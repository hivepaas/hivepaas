import type { ESecuritySettings, EUserRole } from "@application/shared/enums";

export interface User {
    id: string;
    email: string;
    role: EUserRole;
    status: "active" | "pending" | "disabled";
    fullName: string;
    photo: string | null;
    position: string;
    securityOption: ESecuritySettings;
    createdAt: Date;
    updatedAt: Date | null;
    accessExpireAt: Date | null;
    lastAccess: Date | null;
}
