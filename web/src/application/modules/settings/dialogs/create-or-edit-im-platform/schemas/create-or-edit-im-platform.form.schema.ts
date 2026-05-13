import { z } from "zod";

import { EImServiceKind } from "@application/shared/enums";

export const CreateOrEditImPlatformFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.enum([EImServiceKind.Slack, EImServiceKind.Discord]),
    webhook: z.string().trim().min(1, "Webhook URL is required"),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditImPlatformFormInput = z.input<typeof CreateOrEditImPlatformFormSchema>;
export type CreateOrEditImPlatformFormOutput = z.output<typeof CreateOrEditImPlatformFormSchema>;
