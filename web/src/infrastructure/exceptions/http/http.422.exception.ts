import { HttpException } from "./http.exception";

/**
 * Unprocessable Entity
 */
export class Http422Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 422,
        });
    }
}
