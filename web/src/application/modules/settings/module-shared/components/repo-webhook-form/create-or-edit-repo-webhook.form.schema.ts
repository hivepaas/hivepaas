import { z } from "zod";
import { ERepoWebhookKind } from "~/settings/module-shared/enums";

export const CreateOrEditRepoWebhookFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.nativeEnum(ERepoWebhookKind),
    secret: z.string().trim().max(100, "Secret must be 100 characters or less"),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditRepoWebhookFormInput = z.input<typeof CreateOrEditRepoWebhookFormSchema>;
export type CreateOrEditRepoWebhookFormOutput = z.output<typeof CreateOrEditRepoWebhookFormSchema>;
