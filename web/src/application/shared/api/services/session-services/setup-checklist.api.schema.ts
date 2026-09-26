import { z } from "zod";

import type { SetupChecklist, SetupChecklistItem } from "@application/shared/entities";

export const SetupChecklistItemSchema = z
    .object({
        status: z.enum(["todo", "obtaining", "failed", "done"]).catch("todo"),
        domain: z.string().nullish(),
        error: z.string().nullish(),
    })
    .transform(
        (item): SetupChecklistItem => ({
            status: item.status,
            domain: item.domain ?? "",
            error: item.error ?? "",
        }),
    );

export const SetupChecklistSchema = z
    .object({
        dashboardCert: SetupChecklistItemSchema,
        twoFactor: SetupChecklistItemSchema,
        githubApp: SetupChecklistItemSchema,
    })
    .transform((checklist): SetupChecklist => checklist);
