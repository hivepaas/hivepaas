import { HttpException } from "@infrastructure/exceptions/http";

/**
 * Errors the confirm dialog explains better than a toast can.
 *
 * "Too early" is not a failure at all - it is the new configuration still coming
 * up, and the dialog answers it by waiting and retrying. A toast would turn a
 * normal step of the flow into something that looks broken. The other two mean
 * the trial is over, which the dialog says in a full sentence rather than a
 * three-second popup.
 */
const HANDLED_BY_CALLER = new Set([
    "ERR_SETTINGS_CONFIRM_TOO_EARLY",
    "ERR_SETTINGS_CHANGE_SUPERSEDED",
    "ERR_SETTINGS_NO_PENDING_CHANGE",
]);

export function isProbationErrorHandledByCaller(error: Error): boolean {
    return error instanceof HttpException && HANDLED_BY_CALLER.has(error.code);
}
