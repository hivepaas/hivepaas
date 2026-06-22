import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import type { Control, FieldErrors } from "react-hook-form";

import { EAcmeDnsProviderKind } from "@application/shared/enums";

import type { CreateOrEditAcmeDnsProviderFormInput } from "../create-or-edit-acme-dns-provider.form.schema";

import { AcmeDnsProviderField } from "./acme-dns-provider-field.com";

export function AcmeDnsProviderFields({ control, errors, kind }: Props) {
    switch (kind) {
        case EAcmeDnsProviderKind.AcmeDNS:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="acmeDnsApiBase"
                        label="API Base"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="acmeDnsAllowList"
                        label="Allow List"
                        type="textarea"
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="acmeDnsStoragePath"
                        label="Storage Path"
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="acmeDnsStorageBaseUrl"
                        label="Storage Base URL"
                    />
                </>
            );
        case EAcmeDnsProviderKind.Azure:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="azureClientId"
                        label="Client ID"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="azureClientSecret"
                        label="Client Secret"
                        type="password"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="azureSubscriptionId"
                        label="Subscription ID"
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="azureTenantId"
                        label="Tenant ID"
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="azureResourceGroupName"
                        label="Resource Group"
                    />
                </>
            );
        case EAcmeDnsProviderKind.BaiduCloud:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="baiduCloudAccessKey"
                        label="Access Key"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="baiduCloudSecretKey"
                        label="Secret Key"
                        type="password"
                        isRequired
                    />
                </>
            );
        case EAcmeDnsProviderKind.Cloudflare:
            return (
                <>
                    <div className={cn(dashedBorderBox, "text-center text-sm py-2")}>
                        Permissions <span className="text-orange-500">Zone:Read</span> and{" "}
                        <span className="text-orange-500">DNS:Edit</span> are required
                    </div>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="cloudflareAuthToken"
                        label="Auth Token"
                        type="password"
                        isRequired
                    />
                </>
            );
        case EAcmeDnsProviderKind.DigitalOcean:
            return (
                <AcmeDnsProviderField
                    control={control}
                    errors={errors}
                    name="digitalOceanAuthToken"
                    label="Auth Token"
                    type="password"
                    isRequired
                />
            );
        case EAcmeDnsProviderKind.GCloud:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="gCloudServiceAccount"
                        label="Service Account"
                        type="textarea"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="gCloudProjectId"
                        label="Project ID"
                    />
                </>
            );
        case EAcmeDnsProviderKind.GoDaddy:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="goDaddyApiKey"
                        label="API Key"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="goDaddyApiSecret"
                        label="API Secret"
                        type="password"
                        isRequired
                    />
                </>
            );
        case EAcmeDnsProviderKind.Hetzner:
            return (
                <AcmeDnsProviderField
                    control={control}
                    errors={errors}
                    name="hetznerApiToken"
                    label="API Token"
                    type="password"
                    isRequired
                />
            );
        case EAcmeDnsProviderKind.HuaweiCloud:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="huaweiCloudAccessKey"
                        label="Access Key"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="huaweiCloudSecretKey"
                        label="Secret Key"
                        type="password"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="huaweiCloudRegion"
                        label="Region"
                    />
                </>
            );
        case EAcmeDnsProviderKind.Namecheap:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="namecheapApiUser"
                        label="API User"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="namecheapApiKey"
                        label="API Key"
                        type="password"
                        isRequired
                    />
                </>
            );
        case EAcmeDnsProviderKind.RFC2136:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="rfc2136Nameserver"
                        label="Nameserver"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="rfc2136TsigKeyName"
                        label="Tsig Key Name"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="rfc2136TsigSecret"
                        label="Tsig Secret"
                        type="password"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="rfc2136TsigAlgorithm"
                        label="Tsig Algorithm"
                    />
                </>
            );
        case EAcmeDnsProviderKind.Route53:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="route53AccessKeyId"
                        label="Access Key ID"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="route53SecretAccessKey"
                        label="Secret Access Key"
                        type="password"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="route53HostedZoneId"
                        label="Hosted Zone ID"
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="route53Region"
                        label="Region"
                    />
                </>
            );
        case EAcmeDnsProviderKind.TencentCloud:
            return (
                <>
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="tencentCloudSecretId"
                        label="Secret ID"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="tencentCloudSecretKey"
                        label="Secret Key"
                        type="password"
                        isRequired
                    />
                    <AcmeDnsProviderField
                        control={control}
                        errors={errors}
                        name="tencentCloudRegion"
                        label="Region"
                    />
                </>
            );
    }
}

interface Props {
    control: Control<CreateOrEditAcmeDnsProviderFormInput>;
    errors: FieldErrors<CreateOrEditAcmeDnsProviderFormInput>;
    kind: EAcmeDnsProviderKind;
}
