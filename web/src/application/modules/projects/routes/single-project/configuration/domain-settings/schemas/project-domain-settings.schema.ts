import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

const AllowedDomainsSchema = z
    .array(
        z.object({
            value: z.string().transform(value => value.trim()),
        }),
    )
    .superRefine((items, ctx) => {
        const seen = new Set<string>();

        items.forEach((item, index) => {
            const domain = item.value;
            if (!domain) {
                return;
            }

            if (seen.has(domain)) {
                ctx.addIssue({
                    code: "custom",
                    message: `"${domain}" already exists`,
                    path: [index],
                });
                return;
            }

            seen.add(domain);
        });
    });

const CertTypeSchema = z.enum([ESslCertType.LetsEncrypt, ESslCertType.Custom]);
export const ProjectDomainSettingsKeyTypeUnspecified = "__unspecified" as const;

const KeyTypeSchema = z.union([z.nativeEnum(ESslKeyType), z.literal(ProjectDomainSettingsKeyTypeUnspecified)]);

export const ProjectDomainSettingsFormSchema = z.object({
    allowedDomains: AllowedDomainsSchema,
    certSettings: z.object({
        certType: CertTypeSchema,
        email: z.string().transform(value => value.trim()),
        keyType: KeyTypeSchema,
    }),
});

export type ProjectDomainSettingsFormSchemaInput = z.input<typeof ProjectDomainSettingsFormSchema>;
export type ProjectDomainSettingsFormSchemaOutput = z.output<typeof ProjectDomainSettingsFormSchema>;

export const emptyProjectDomainSettingsFormDefaults: ProjectDomainSettingsFormSchemaInput = {
    allowedDomains: [],
    certSettings: {
        certType: ESslCertType.LetsEncrypt,
        email: "",
        keyType: ESslKeyType.ECP256,
    },
};
