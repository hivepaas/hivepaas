import type { ProjectApps_Copy_Req, ProjectApps_PrepareCopy_Res } from "~/projects/api/services";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import type { CopyProjectAppFormInput, CopyProjectAppFormOutput } from "../schemas";

type CopyProjectAppPayload = Omit<ProjectApps_Copy_Req["data"], "projectID" | "appID">;

export function mapPreparedCopyToFormInput(data: ProjectApps_PrepareCopy_Res["data"]): CopyProjectAppFormInput {
    return {
        ...data,
        targetStatus:
            data.targetStatus === EProjectAppStatus.Disabled ? EProjectAppStatus.Disabled : EProjectAppStatus.Active,
        copyHttpSettings: {
            copy: data.copyHttpSettings.copy,
            copyDomainSettings: data.copyHttpSettings.copyDomainSettings.map(domain => ({
                sourceDomain: domain.sourceDomain,
                targetDomain: domain.targetDomain,
                sourceSslCert: domain.sourceSslCert,
                targetSslCert: domain.targetSslCert,
            })),
        },
    };
}

export function mapCopyProjectAppFormToPayload(values: CopyProjectAppFormOutput): CopyProjectAppPayload {
    return {
        ...values,
        copyHttpSettings: {
            copy: values.copyHttpSettings.copy,
            copyDomainSettings: values.copyHttpSettings.copyDomainSettings
                .filter(domain => domain.targetDomain.trim() !== "")
                .map(domain => ({
                    sourceDomain: domain.sourceDomain,
                    targetDomain: domain.targetDomain.trim(),
                    sourceSslCert: {
                        id: domain.sourceSslCert?.id ?? "",
                    },
                    targetSslCert: {
                        id: domain.targetSslCert?.id ?? "",
                    },
                })),
        },
    };
}
