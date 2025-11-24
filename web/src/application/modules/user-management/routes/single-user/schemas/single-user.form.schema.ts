import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

export const SingleUserFormSchema = z.object({
    fullName: z
        .string()
        .trim()
        .min(1, "Full Name is required")
        .max(100, "Full Name should be at most 100 characters long"),
    email: z.string(),
    username: z
        .string()
        .trim()
        .max(100, "Username should be at most 100 characters long")
        .regex(/^[a-zA-Z0-9]*$/, "Username must contain only letters and numbers"),
    position: z.string(),
    notes: z.string().optional(),
    role: z.nativeEnum(EUserRole),
    accessExpireAt: z.date().nullable(),
    securityOption: z.nativeEnum(ESecuritySettings),
    projectAccesses: z.array(AccessSchema),
    moduleAccesses: z.array(AccessSchema),
});

export type SingleUserFormSchemaInput = z.input<typeof SingleUserFormSchema>;
export type SingleUserFormSchemaOutput = z.output<typeof SingleUserFormSchema>;
