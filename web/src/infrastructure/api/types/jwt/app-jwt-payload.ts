import { type JwtPayload } from "jwt-decode";

export type AppJwtPayload = JwtPayload & {
    userId: string;
};
