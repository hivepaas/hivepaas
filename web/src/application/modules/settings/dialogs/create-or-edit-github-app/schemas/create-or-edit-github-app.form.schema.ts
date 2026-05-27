import { z } from "zod";

export const CreateOrEditGithubAppFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    organization: z.string().trim(),
    appId: z.coerce.number().int("App ID must be an integer").positive("App ID is required"),
    installationId: z.coerce.number().int("Installation ID must be an integer").positive("Installation ID is required"),
    clientId: z.string().trim().min(1, "Client ID is required"),
    clientSecret: z.string().min(1, "Client secret is required"),
    privateKey: z.string().trim().min(1, "Private key is required"),
    ssoEnabled: z.boolean(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditGithubAppFormInput = z.input<typeof CreateOrEditGithubAppFormSchema>;
export type CreateOrEditGithubAppFormOutput = z.output<typeof CreateOrEditGithubAppFormSchema>;
