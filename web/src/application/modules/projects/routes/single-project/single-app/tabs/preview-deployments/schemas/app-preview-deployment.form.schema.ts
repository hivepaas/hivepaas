import { z } from "zod";

export const PREVIEW_DEPLOYMENT_TRIGGER = {
    PullRequestComments: "pull-request-comments",
    DashboardUI: "dashboard-ui",
} as const;

export type PreviewDeploymentTrigger = (typeof PREVIEW_DEPLOYMENT_TRIGGER)[keyof typeof PREVIEW_DEPLOYMENT_TRIGGER];

const PreviewDeploymentTriggerSchema = z.union([
    z.literal(PREVIEW_DEPLOYMENT_TRIGGER.PullRequestComments),
    z.literal(PREVIEW_DEPLOYMENT_TRIGGER.DashboardUI),
]);

export const AppPreviewDeploymentFormSchema = z
    .object({
        trigger: PreviewDeploymentTriggerSchema,
        repoRef: z.string().trim().max(255, "Git Ref must be 255 characters or fewer"),
        customSubdomain: z.string().trim().max(63, "Custom Subdomain must be 63 characters or fewer"),
        noStart: z.boolean(),
    })
    .superRefine((values, ctx) => {
        if (values.trigger !== PREVIEW_DEPLOYMENT_TRIGGER.DashboardUI) {
            return;
        }

        if (!values.repoRef) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["repoRef"],
                message: "Git Ref is required",
            });
        }
    });

export type AppPreviewDeploymentFormInput = z.input<typeof AppPreviewDeploymentFormSchema>;
export type AppPreviewDeploymentFormOutput = z.output<typeof AppPreviewDeploymentFormSchema>;

export const DEFAULT_APP_PREVIEW_DEPLOYMENT_FORM_VALUES: AppPreviewDeploymentFormInput = {
    trigger: PREVIEW_DEPLOYMENT_TRIGGER.PullRequestComments,
    repoRef: "",
    customSubdomain: "",
    noStart: false,
};
