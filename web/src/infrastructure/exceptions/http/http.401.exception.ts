import { HttpException } from "./http.exception";

/**
 * Unauthorized
 */
export class Http401Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 401,
        });
    }
}
