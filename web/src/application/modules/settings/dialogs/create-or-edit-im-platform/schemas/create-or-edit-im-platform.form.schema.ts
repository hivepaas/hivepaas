import { z } from "zod";

import { EImServiceKind } from "@application/shared/enums";

export const CreateOrEditImPlatformFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        kind: z.enum([EImServiceKind.Slack, EImServiceKind.Discord, EImServiceKind.Telegram]),
        webhook: z.string().trim(),
        botToken: z.string().trim(),
        chatId: z.string().trim(),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((values, ctx) => {
        if (values.kind !== EImServiceKind.Telegram && !values.webhook) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["webhook"],
                message: "Webhook URL is required",
            });
        }

        if (values.kind === EImServiceKind.Telegram && !values.botToken) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["botToken"],
                message: "Bot Token is required",
            });
        }

        if (values.kind === EImServiceKind.Telegram && !values.chatId) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["chatId"],
                message: "Chat ID is required",
            });
        }
    });

export type CreateOrEditImPlatformFormInput = z.input<typeof CreateOrEditImPlatformFormSchema>;
export type CreateOrEditImPlatformFormOutput = z.output<typeof CreateOrEditImPlatformFormSchema>;
