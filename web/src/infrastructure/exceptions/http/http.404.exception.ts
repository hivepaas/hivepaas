import { HttpException } from "./http.exception";

/**
 * Not found
 */
export class Http404Exception extends HttpException {
    constructor(message: string) {
        super({
            message,
            status: 404,
        });
    }
}
