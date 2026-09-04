import type { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface UserBase {
    id: string;
    email: string;
    role: EUserRole;
    status: OpenApiConstant<EUserStatus>;
    fullName: string;
    username: string;
    photo: string | null;
    position: string;
    securityOption: ESecuritySettings;
    createdAt: Date;
    updatedAt: Date | null;
    accessExpireAt: Date | null;
    lastAccess: Date | null;
    /**
     * Access the user has on each project, broken down per env. Permissions are
     * granted per env only; the project level is just a grouping.
     */
    projectAccesses: {
        id: string;
        name: string;
        envAccesses: {
            id: string;
            name: string;
            color: string;
            access: {
                read: boolean;
                execute: boolean;
                write: boolean;
                delete: boolean;
            };
        }[];
    }[];
    notes?: string;
    moduleAccesses: {
        id: string;
        name: string;
        access: {
            read: boolean;
            execute: boolean;
            write: boolean;
            delete: boolean;
        };
    }[];
}
