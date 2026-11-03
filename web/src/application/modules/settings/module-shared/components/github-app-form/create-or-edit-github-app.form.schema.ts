import { z } from "zod";

export const CreateOrEditGithubAppFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        organization: z.string().trim(),
        appId: z.coerce.number().int("App ID must be an integer").positive("App ID is required"),
        installationId: z.coerce
            .number()
            .int("Installation ID must be an integer")
            .positive("Installation ID is required"),
        clientId: z.string().trim(),
        clientSecret: z.string(),
        privateKey: z.string().trim().min(1, "Private key is required"),
        ssoEnabled: z.boolean(),
        inheritable: z.boolean(),
        default: z.boolean(),
    })
    // The client is OAuth: only signing in through the app (SSO) needs it.
    .superRefine((values, ctx) => {
        if (!values.ssoEnabled) {
            return;
        }
        if (values.clientId === "") {
            ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["clientId"], message: "Client ID is required for SSO" });
        }
        if (values.clientSecret === "") {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["clientSecret"],
                message: "Client secret is required for SSO",
            });
        }
    });

export type CreateOrEditGithubAppFormInput = z.input<typeof CreateOrEditGithubAppFormSchema>;
export type CreateOrEditGithubAppFormOutput = z.output<typeof CreateOrEditGithubAppFormSchema>;
