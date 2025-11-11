import type { ESecuritySettings } from "../enums";

export interface Profile {
    /**
     * User ID
     * @example "8f7123fd-1862-40ad-a082-d16d421c7d8b"
     */
    id: string;

    /**
     * Username
     * @example "johndoe"
     */
    username: string;

    /**
     * Full name
     * @example "John Doe"
     */
    fullName?: string;

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
}
