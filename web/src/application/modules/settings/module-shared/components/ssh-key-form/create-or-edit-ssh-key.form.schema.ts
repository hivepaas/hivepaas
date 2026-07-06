import { z } from "zod";

import { ESSHKeyKind, ESSHKeyType } from "@application/shared/enums";

export const CreateOrEditSSHKeyFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.nativeEnum(ESSHKeyKind),
    keyType: z.union([z.nativeEnum(ESSHKeyType), z.literal("")]),
    publicKey: z.string(),
    privateKey: z.string().min(1, "Private key is required"),
    passphrase: z.string(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditSSHKeyFormInput = z.input<typeof CreateOrEditSSHKeyFormSchema>;
export type CreateOrEditSSHKeyFormOutput = z.output<typeof CreateOrEditSSHKeyFormSchema>;
