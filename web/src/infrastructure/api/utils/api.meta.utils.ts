import { z } from "zod";

const MetaNotificationsApiSchema = z.object({
    meta: z
        .object({
            message: z.string().optional(),
            warning: z.string().optional(),
            error: z.string().optional(),
        })
        .optional()
        .nullable(),
});

export interface MetaNotifications {
    message?: string;
    warning?: string;
    error?: string;
}

/**
 * Safely extracts notification fields from a raw API response body.
 * Returns null when meta is absent or all fields are empty/whitespace.
 * Never throws.
 */
export function extractMetaNotifications(responseData: unknown): MetaNotifications | null {
    const parsed = MetaNotificationsApiSchema.safeParse(responseData);

    if (!parsed.success || !parsed.data.meta) {
        return null;
    }

    const { message, warning, error } = parsed.data.meta;

    const result: MetaNotifications = {};

    const trimmedMessage = message?.trim();
    const trimmedWarning = warning?.trim();
    const trimmedError = error?.trim();

    if (trimmedMessage) result.message = trimmedMessage;
    if (trimmedWarning) result.warning = trimmedWarning;
    if (trimmedError) result.error = trimmedError;

    if (!result.message && !result.warning && !result.error) {
        return null;
    }

    return result;
}
