import { AxiosError } from "axios";

/**
 * Cancel exception
 */
export class CancelException extends Error {
    constructor(message: string) {
        super(message);

        this.code = AxiosError.ERR_CANCELED;
    }

    readonly code: string;
}
