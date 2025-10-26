import { HttpException } from "./http.exception";

/**
 * Forbidden
 */
export class Http403Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 403,
        });
    }
}
