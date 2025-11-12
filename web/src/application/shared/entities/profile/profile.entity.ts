import type { ESecuritySettings } from "@application/shared/enums";

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
    fullName?: string | null;

    /**
     * Photo
     * @example "https://example.com/photo.jpg"
     */
    photo?: string | null;

    /**
     * Email
     * @example "john.dou@.example.com"
     */
    email?: string | null;

    securityOption: ESecuritySettings;

    mfaSecret: string;

    username: string;
}
