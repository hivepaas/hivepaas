import { z } from "zod";

export const CreateOrEditNotificationTargetFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        emailEnabled: z.boolean(),
        senderEmailAccountId: z.string().trim(),
        notifyAdmins: z.boolean(),
        notifyProjectOwners: z.boolean(),
        notifyProjectMembers: z.boolean(),
        customAddresses: z.string().trim(),
        slackEnabled: z.boolean(),
        slackWebhookId: z.string().trim(),
        discordEnabled: z.boolean(),
        discordWebhookId: z.string().trim(),
        minSendInterval: z.string().trim().min(1, "Override Min Send Interval is required"),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((values, ctx) => {
        if (values.emailEnabled && !values.senderEmailAccountId) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["senderEmailAccountId"],
                message: "Sender Email Account is required",
            });
        }
        if (values.slackEnabled && !values.slackWebhookId) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["slackWebhookId"],
                message: "Webhook is required",
            });
        }
        if (values.discordEnabled && !values.discordWebhookId) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["discordWebhookId"],
                message: "Webhook is required",
            });
        }
    });

export type CreateOrEditNotificationTargetFormInput = z.input<typeof CreateOrEditNotificationTargetFormSchema>;
export type CreateOrEditNotificationTargetFormOutput = z.output<typeof CreateOrEditNotificationTargetFormSchema>;
