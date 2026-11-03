import { AxiosError } from "axios";

export class RouteNotFoundException extends Error {
    constructor(route: string) {
        super(`404 Route not found: ${route}`);

        this.code = AxiosError.ERR_BAD_REQUEST;
    }

    readonly code: string;
}
