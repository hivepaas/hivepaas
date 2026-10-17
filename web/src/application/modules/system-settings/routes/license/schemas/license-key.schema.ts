import { z } from "zod";

/** The longest key text the server takes: license.MaxKeyTextLen in the backend. */
const MAX_KEY_LENGTH = 32 * 1024;

export const LicenseKeyFormSchema = z.object({
    // Whitespace is dropped first, as the server drops it (license.NormalizeKeyText):
    // a key wrapped by an email client is the same key, and the length checked is
    // the key's own.
    key: z
        .string()
        .transform(value => value.replace(/\s+/gu, ""))
        .pipe(z.string().min(1, "Paste a license key").max(MAX_KEY_LENGTH, "This is longer than any license key")),
});

export type LicenseKeyFormInput = z.input<typeof LicenseKeyFormSchema>;
export type LicenseKeyFormOutput = z.output<typeof LicenseKeyFormSchema>;
