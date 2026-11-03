import { HttpException } from "@infrastructure/exceptions/http";

/**
 * Refusals the import panel answers in place rather than with a popup: it asks
 * for the passphrase, or shows the new plan to confirm again.
 */
const HANDLED_BY_CALLER = new Set([
    "ERR_SPEC_PASSPHRASE_REQUIRED",
    "ERR_SPEC_PASSPHRASE_INVALID",
    "ERR_SPEC_IMPORT_PLAN_CHANGED",
]);

export function isSpecImportErrorHandledByCaller(error: Error): boolean {
    return error instanceof HttpException && HANDLED_BY_CALLER.has(error.code);
}

export function specImportErrorCode(error: unknown): string | undefined {
    return error instanceof HttpException ? error.code : undefined;
}
