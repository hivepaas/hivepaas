import { z } from "zod";
import { MANAGED_RESTORE_TARGET } from "~/projects/domain";
import { formatManagedTime } from "~/projects/module-shared/utils";

const APP_NAME_MAX_LENGTH = 100;

const ManagedRestoreFormBaseSchema = z.object({
    kind: z.nativeEnum(MANAGED_RESTORE_TARGET),
    time: z.date().nullable(),
    point: z.string(),
    backup: z.string(),
    dataVolume: z.string().min(1, "Choose the volume the new app keeps its data on"),
    name: z.string().trim().max(APP_NAME_MAX_LENGTH, `At most ${APP_NAME_MAX_LENGTH} characters`),
});

export type ManagedRestoreFormInput = z.input<typeof ManagedRestoreFormBaseSchema>;
export type ManagedRestoreFormOutput = z.output<typeof ManagedRestoreFormBaseSchema>;

/** A restore's target is checked against the window it must be inside, as the source's status gives it. */
export function createManagedRestoreFormSchema(
    restoreWindow: { from: Date; to: Date } | undefined,
): z.ZodType<ManagedRestoreFormOutput, z.ZodTypeDef, ManagedRestoreFormInput> {
    return ManagedRestoreFormBaseSchema.superRefine((value, ctx) => {
        const issue = (message: string, path: string[]) => {
            ctx.addIssue({ code: z.ZodIssueCode.custom, message, path });
        };

        if (value.kind === MANAGED_RESTORE_TARGET.time) {
            if (!value.time) {
                issue("Choose the time to restore to", ["time"]);
            } else if (restoreWindow && (value.time < restoreWindow.from || value.time > restoreWindow.to)) {
                issue(`Between ${formatManagedTime(restoreWindow.from)} and ${formatManagedTime(restoreWindow.to)}`, [
                    "time",
                ]);
            }
        }
        if (value.kind === MANAGED_RESTORE_TARGET.point && value.point === "") {
            issue("Choose the restore point", ["point"]);
        }
        if (value.kind === MANAGED_RESTORE_TARGET.backup && value.backup === "") {
            issue("Choose the base backup", ["backup"]);
        }
    });
}
