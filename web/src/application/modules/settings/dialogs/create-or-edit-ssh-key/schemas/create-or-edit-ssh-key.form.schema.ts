import { z } from "zod";

export const CreateOrEditSSHKeyFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    privateKey: z.string().min(1, "Private key is required"),
    passphrase: z.string(),
    targets: z.array(z.object({ value: z.string().trim().min(1, "Target is required") })),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditSSHKeyFormInput = z.input<typeof CreateOrEditSSHKeyFormSchema>;
export type CreateOrEditSSHKeyFormOutput = z.output<typeof CreateOrEditSSHKeyFormSchema>;
