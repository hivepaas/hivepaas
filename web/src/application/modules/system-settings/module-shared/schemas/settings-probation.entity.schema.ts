import { z } from "zod";

/**
 * Shared by every settings kind that can be put on trial - routing settings and
 * HivePaaS service settings today. One schema because one server type produces
 * it: two copies would let the two endpoints drift into disagreeing about what a
 * pending change looks like.
 */
export const SettingsPendingChangeSchema = z.object({
    changeId: z.string(),
    appliedAt: z.coerce.date(),
    confirmableFrom: z.coerce.date(),
    deadlineAt: z.coerce.date(),
});
