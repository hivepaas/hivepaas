import type { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

export interface UserBase {
    id: string;
    email: string;
    role: EUserRole;
    status: EUserStatus;
    fullName: string;
    username: string;
    photo: string | null;
    position: string;
    securityOption: ESecuritySettings;
    createdAt: Date;
    updatedAt: Date | null;
    accessExpireAt: Date | null;
    lastAccess: Date | null;
    projectAccesses: {
        id: string;
        name: string;
        access: {
            read: boolean;
            write: boolean;
            delete: boolean;
        };
    }[];
    notes?: string;
    moduleAccesses: {
        id: string;
        name: string;
        access: {
            read: boolean;
            write: boolean;
            delete: boolean;
        };
    }[];
}
