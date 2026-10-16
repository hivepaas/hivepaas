import { Http400Exception, HttpException } from "@infrastructure/exceptions/http";

/**
 * Refusals of a pasted key that the license form answers beside the box rather
 * than with a popup: a key that does not verify, one for another installation -
 * whose message names this installation's ID, the one thing the person needs
 * next - and one past its grace. The stored key is kept in each case.
 */
const KEY_REFUSALS = new Set(["ERR_LICENSE_INVALID", "ERR_LICENSE_OTHER_INSTALLATION", "ERR_LICENSE_EXPIRED"]);

/** The field a 400 names when it is about the key: missing, or too long to be one. */
const KEY_FIELD = "key";

/**
 * What the license form says beside the key box about a refusal: the server's own
 * sentence, such as "This license key is for another installation. This
 * installation's ID is …". Undefined for any other failure, which is toasted.
 */
export function licenseKeyRefusalMessage(error: unknown): string | undefined {
    if (!(error instanceof HttpException)) {
        return undefined;
    }

    if (KEY_REFUSALS.has(error.code)) {
        return error.message;
    }

    // The detail of a 400 only says the input was invalid; the field's own
    // message says how.
    if (error instanceof Http400Exception) {
        const keyError = error.errors.find(item => item.path === KEY_FIELD);

        if (keyError) {
            return keyError.message || error.message;
        }
    }

    return undefined;
}

export function isLicenseKeyRefusal(error: unknown): boolean {
    return licenseKeyRefusalMessage(error) !== undefined;
}
