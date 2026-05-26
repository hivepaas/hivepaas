import type { EUserRole } from "@application/shared/enums";

export interface UserPublic {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string | null;
    role: EUserRole;
}
