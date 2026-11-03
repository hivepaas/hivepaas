import { AxiosError } from "axios";

/**
 * Network exception
 */
export class NetworkException extends Error {
    constructor(message: string) {
        super(message);

        this.code = AxiosError.ERR_NETWORK;
    }

    readonly code: string;
}
