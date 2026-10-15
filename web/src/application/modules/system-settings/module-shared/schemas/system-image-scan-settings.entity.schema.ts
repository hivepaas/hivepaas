import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

const NotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const text = z
    .string()
    .nullish()
    .transform(value => value ?? "");

const date = z.coerce
    .date()
    .nullish()
    .transform(value => value ?? null);

export const SystemImageScanSettingsEntitySchema = SettingsBaseEntitySchema.omit({
    description: true,
}).extend({
    type: z.string(),
    enabled: z.boolean().catch(false),
    volume: z
        .object({ id: z.string() })
        .nullish()
        .transform(value => value ?? null),
    dbRepository: text,
    javaDbRepository: text,
    schedule: z
        .object({
            interval: z.string().nullish(),
            initialTime: z.coerce.date().nullish(),
        })
        .transform(({ interval, initialTime }) => ({
            interval: interval ?? "1d",
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
    cpuLimit: z.number().catch(1),
    memoryLimit: z.string().catch("512mb"),
    nextRuns: z
        .array(z.coerce.date())
        .nullish()
        .transform(value => value ?? []),
    scannerStatus: z
        .object({
            volume: z
                .object({
                    id: z.string(),
                    name: text,
                    missing: z.boolean().catch(false),
                    nodeId: text,
                    nodeName: text,
                    nodeLabel: text,
                    shared: z.boolean().catch(false),
                })
                .nullish()
                .transform(value => value ?? null),
            image: text,
            trivyVersion: text,
            dbUpdatedAt: date,
            lastScanAt: date,
            lastError: text,
        })
        .nullish()
        .transform(value => value ?? null),
});

const CountSchema = z.object({ total: z.number().catch(0), fixable: z.number().catch(0) });

export const ImageScanAppSchema = z.object({
    app: z.object({ id: z.string(), name: z.string() }),
    project: z.object({ id: z.string(), name: z.string() }),
    projectEnv: z.object({ id: z.string(), name: z.string() }),
    scannedAt: z.coerce.date(),
    error: text,
    summary: z.object({
        critical: CountSchema,
        high: CountSchema,
        medium: CountSchema,
        low: CountSchema,
        unknown: CountSchema,
        ignored: z.number().catch(0),
    }),
    images: z
        .array(
            z.object({
                imageId: z.string(),
                imageRef: text,
                osFamily: text,
                osName: text,
                osEosl: z.boolean().catch(false),
                status: z.string(),
                error: text,
                scannedAt: z.coerce.date(),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
});
