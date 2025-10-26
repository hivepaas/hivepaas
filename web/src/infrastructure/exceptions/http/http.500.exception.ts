import { HttpException } from "./http.exception";

/**
 * Internal server error
 */
export class Http500Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 500,
        });
    }
}
