import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

export const InviteUserFormSchema = z.object({
    email: z.string().email("Invalid email format").trim().min(1, "Email is required"),
    role: z.nativeEnum(EUserRole),
    securityOption: z.nativeEnum(ESecuritySettings),
    accessExpireAt: z.date().nullable(),
    projectAccesses: z.array(AccessSchema),
    moduleAccesses: z.array(AccessSchema),
});

export type InviteUserFormInput = z.input<typeof InviteUserFormSchema>;
export type InviteUserFormOutput = z.output<typeof InviteUserFormSchema>;
