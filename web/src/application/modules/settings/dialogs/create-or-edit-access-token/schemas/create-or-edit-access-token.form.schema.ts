import { z } from "zod";

import { EAccessTokenKind } from "@application/shared/enums";

export const CreateOrEditAccessTokenFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.nativeEnum(EAccessTokenKind),
    user: z.string().trim().min(1, "User is required"),
    token: z.string().min(1, "Token is required"),
    baseURL: z.string().trim().min(1, "Base URL is required"),
    expireAt: z.date().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditAccessTokenFormInput = z.input<typeof CreateOrEditAccessTokenFormSchema>;
export type CreateOrEditAccessTokenFormOutput = z.output<typeof CreateOrEditAccessTokenFormSchema>;
