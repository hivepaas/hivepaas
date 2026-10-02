import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

const NotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

export const SystemRegistryAuthRenewalSettingsEntitySchema = SettingsBaseEntitySchema.omit({
    description: true,
}).extend({
    type: z.string(),
    schedule: z
        .object({
            interval: z.string().nullish(),
            initialTime: z.coerce.date().nullish(),
        })
        .transform(({ interval, initialTime }) => ({
            interval: interval ?? "",
            initialTime: initialTime ?? null,
        })),
    notification: z
        .object({
            success: NotificationRefSchema.nullish().transform(value => value ?? undefined),
            successUseDefault: z.boolean(),
            failure: NotificationRefSchema.nullish().transform(value => value ?? undefined),
            failureUseDefault: z.boolean(),
        })
        .nullish()
        .transform(value => value ?? null),
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
});
