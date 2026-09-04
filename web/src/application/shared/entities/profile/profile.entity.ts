import type { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";
import type { ModulePermission, ProjectPermission } from "@application/shared/permissions/types";

import type { OpenApiConstant } from "@infrastructure/api";

export interface Profile {
    /**
     * User ID
     * @example "8f7123fd-1862-40ad-a082-d16d421c7d8b"
     */
    id: string;

    /**
     * Full name
     * @example "John Doe"
     */
    fullName: string | null;

    /**
     * Photo
     * @example "https://example.com/photo.jpg"
     */
    photo: string | null;

    /**
     * Email
     * @example "john.dou@.example.com"
     */
    email: string | null;

    securityOption: ESecuritySettings;

    mfaSecret: string;

    username: string;

    position: string;

    notes: string;

    role: EUserRole;

    status: OpenApiConstant<EUserStatus>;

    lastAccess: Date | null;

    accessExpireAt: Date | null;

    createdAt: Date;

    /**
     * Access per project. `access` is the union across the project's envs; read
     * `envAccesses` when a check must be exact for one env.
     */
    projectAccesses: {
        id: string;
        name: string;
        access: {
            read: boolean;
            execute: boolean;
            write: boolean;
            delete: boolean;
        };
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

    modulePermissions: ModulePermission[];

    projectPermissions: ProjectPermission[];

    mfaTotpActivated?: boolean;
}
