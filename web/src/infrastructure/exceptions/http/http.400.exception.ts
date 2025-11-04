import { HttpException } from "./http.exception";

/**
 * Bad request
 */
export class Http400Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 400,
        });
    }
}
