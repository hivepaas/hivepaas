import { z } from "zod";
import { EGithubAppOwnerType } from "~/settings/module-shared/enums";

export const ProvisionGithubAppFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        ownerType: z.enum([EGithubAppOwnerType.Organization, EGithubAppOwnerType.User]),
        org: z.string().trim(),
        ssoEnabled: z.boolean(),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((value, ctx) => {
        if (value.ownerType === EGithubAppOwnerType.Organization && !value.org) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["org"],
                message: "Organization is required",
            });
        }
    });

export type ProvisionGithubAppFormInput = z.input<typeof ProvisionGithubAppFormSchema>;
export type ProvisionGithubAppFormOutput = z.output<typeof ProvisionGithubAppFormSchema>;
