import { HttpException } from "./http.exception";

/**
 * Conflict
 */
export class Http409Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 409,
        });
    }
}
