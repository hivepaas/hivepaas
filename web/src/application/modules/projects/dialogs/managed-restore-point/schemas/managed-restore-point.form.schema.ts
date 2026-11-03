import { z } from "zod";
import { MAX_POINT_KEEP_DAYS, POINT_NAME_PATTERN, RESERVED_POINT_PREFIX } from "~/projects/module-shared/utils";

const NOTE_MAX_LENGTH = 500;
const DAY_MS = 24 * 60 * 60 * 1000;

const ManagedRestorePointFormBaseSchema = z.object({
    name: z.string().trim(),
    note: z.string().trim().max(NOTE_MAX_LENGTH, `At most ${NOTE_MAX_LENGTH} characters`),
    keepUntil: z.date().nullable(),
});

export type ManagedRestorePointFormInput = z.input<typeof ManagedRestorePointFormBaseSchema>;
export type ManagedRestorePointFormOutput = z.output<typeof ManagedRestorePointFormBaseSchema>;

/** A point's name is checked only when it is made: after, it is the point's for good. */
export function createManagedRestorePointFormSchema(
    creating: boolean,
): z.ZodType<ManagedRestorePointFormOutput, z.ZodTypeDef, ManagedRestorePointFormInput> {
    return ManagedRestorePointFormBaseSchema.superRefine((value, ctx) => {
        const issue = (message: string, path: string[]) => {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message, path });
        };

        if (creating) {
            if (!POINT_NAME_PATTERN.test(value.name)) {
                issue("Letters, digits, dots, dashes and underscores, starting with a letter or a digit: 63 at most", [
                    "name",
                ]);
            } else if (value.name.startsWith(RESERVED_POINT_PREFIX)) {
                issue(`Names starting with ${RESERVED_POINT_PREFIX} are HivePaaS's own`, ["name"]);
            }
        }
        if (value.keepUntil) {
            const now = Date.now();
            if (value.keepUntil.getTime() <= now) {
                issue("A date to come", ["keepUntil"]);
            } else if (value.keepUntil.getTime() > now + MAX_POINT_KEEP_DAYS * DAY_MS) {
                issue(`At most ${MAX_POINT_KEEP_DAYS} days from now`, ["keepUntil"]);
            }
        }
    });
}
