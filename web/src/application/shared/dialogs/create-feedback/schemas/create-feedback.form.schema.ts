import { z } from "zod";

import { SupportFeedbackCategory } from "@application/shared/enums";

const SupportFeedbackCategorySchema = z.enum([
    SupportFeedbackCategory.General,
    SupportFeedbackCategory.SecurityReport,
    SupportFeedbackCategory.BugIssueReport,
    SupportFeedbackCategory.Licensing,
]);

export const CreateFeedbackFormSchema = z.object({
    category: SupportFeedbackCategorySchema,
    name: z.string(),
    email: z
        .string()
        .trim()
        .refine(value => !value || z.string().email().safeParse(value).success, "Invalid email address"),
    company: z.string(),
    subject: z.string().trim().min(1, "Subject is required"),
    description: z.string().trim().min(1, "Description is required"),
});

export type CreateFeedbackFormSchemaInput = z.input<typeof CreateFeedbackFormSchema>;
export type CreateFeedbackFormSchemaOutput = z.output<typeof CreateFeedbackFormSchema>;

export const emptyCreateFeedbackFormDefaults: CreateFeedbackFormSchemaInput = {
    category: SupportFeedbackCategory.General,
    name: "",
    email: "",
    company: "",
    subject: "",
    description: "",
};
