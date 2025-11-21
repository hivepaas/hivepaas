import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

export const ProfileFormSchema = z.object({
    photo: z.string().nullable(),
    photoUpload: z.union([
        z.object({
            fileName: z.string(),
            dataBase64: z.string(),
        }), // Object with data to upload new photo
        z.null(), // null to keep photo unchanged
    ]),
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
    position: z.string(),
    notes: z.string(),
    role: z.nativeEnum(EUserRole),
    accessExpireAt: z.date().nullable(),
    securityOption: z.nativeEnum(ESecuritySettings),
    projectAccesses: z.array(AccessSchema),
    moduleAccesses: z.array(AccessSchema),
});

export type ProfileFormSchemaInput = z.input<typeof ProfileFormSchema>;
export type ProfileFormSchemaOutput = z.output<typeof ProfileFormSchema>;
