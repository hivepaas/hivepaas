import { AxiosError } from "axios";

/**
 * Timeout exception
 */
export class TimeoutException extends Error {
    constructor(message: string) {
        super(message);

        this.code = AxiosError.ETIMEDOUT;
    }

    readonly code: string;
}
