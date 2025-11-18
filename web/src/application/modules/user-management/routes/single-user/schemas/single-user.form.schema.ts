import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

export const SingleUserFormSchema = z.object({
    fullName: z
        .string()
        .trim()
        .min(1, "Full Name is required")
        .max(255, "Full Name should be at most 255 characters long"),
    email: z.string(),
    username: z
        .string()
        .trim()
        .max(255, "Username should be at most 255 characters long")
        .regex(/^[a-zA-Z0-9]*$/, "Username must contain only letters and numbers"),
    position: z
        .string()
        .trim()
        .min(1, "Position is required")
        .max(255, "Position should be at most 255 characters long"),
    notes: z.string().optional(),
    role: z.nativeEnum(EUserRole),
    accessExpireAt: z.date().nullable(),
    securityOption: z.nativeEnum(ESecuritySettings),
    projectAccess: z.array(AccessSchema),
    moduleAccess: z.array(AccessSchema),
});

export type SingleUserFormSchemaInput = z.input<typeof SingleUserFormSchema>;
export type SingleUserFormSchemaOutput = z.output<typeof SingleUserFormSchema>;
