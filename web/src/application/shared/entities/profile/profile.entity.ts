import type { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

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

    status: EUserStatus;

    lastAccess: Date | null;

    accessExpireAt: Date | null;

    createdAt: Date;

    projectAccesses: {
        id: string;
        name: string;
        access: {
            read: boolean;
            write: boolean;
            delete: boolean;
        };
    }[];

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
