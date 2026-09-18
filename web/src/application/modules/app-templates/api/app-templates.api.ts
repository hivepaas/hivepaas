import { BaseApi, parseApiError } from "@infrastructure/api";

export interface AppTemplateCategory {
    id: string;
    title: string;
    /** How many templates opening this category shows. A parent counts each of
     *  its templates once, however many children it is listed in. */
    count: number;
    children?: AppTemplateCategory[];
}

export interface AppTemplateTag {
    id: string;
    title: string;
}

export interface AppTemplateCatalog {
    source: string;
    revision: string;
    categories: AppTemplateCategory[];
    tags: AppTemplateTag[];
}

export interface AppTemplateVariantSummary {
    name: string;
    default: boolean;
}

export interface AppTemplateVersionSummary {
    name: string;
    release: string;
    default: boolean;
    deprecated: boolean;
    variants: string[];
}

export interface AppTemplateSummary {
    name: string;
    title: string;
    tagline: string;
    categories: string[];
    tags: string[];
    aliases: string[];
    iconUrl: string;
    license?: string;
    variants: AppTemplateVariantSummary[];
    versions: AppTemplateVersionSummary[];
    compatible: boolean;
}

export interface ListAppTemplatesFilter {
    category?: string;
    tag?: string;
    search?: string;
    pageOffset?: number;
    pageLimit?: number;
}

export interface ListAppTemplatesPageMeta {
    offset: number;
    limit: number;
    total: number;
}

export interface ListAppTemplatesMeta {
    page?: ListAppTemplatesPageMeta;
    total?: number;
    count?: number;
    pageOffset?: number;
    pageLimit?: number;
}

export interface ListAppTemplatesResponse {
    meta: ListAppTemplatesMeta;
    data: AppTemplateSummary[];
}

export interface AppTemplateLinks {
    website?: string;
    documentation?: string;
    source?: string;
}

export interface AppTemplateVariant {
    name: string;
    title: string;
    description: string;
    default: boolean;
}

export interface AppTemplateParamOption {
    value: string;
    title: string;
}

export interface AppTemplateParam {
    name: string;
    title: string;
    description: string;
    type: string;
    default?: unknown;
    optional: boolean;
    pattern?: string;
    minLength?: number;
    maxLength?: number;
    min?: unknown;
    max?: unknown;
    generated: boolean;
    options?: AppTemplateParamOption[];
}

export interface AppTemplateDependency {
    name: string;
    title: string;
    template: string;
    templateTitle?: string;
    version?: string;
    variant?: string;
    parameters?: AppTemplateParam[];
}

export interface AppTemplateDetail {
    source: string;
    revision: string;
    name: string;
    title: string;
    tagline: string;
    description: string;
    categories: string[];
    tags: string[];
    iconUrl: string;
    links?: AppTemplateLinks;
    license: string;
    compatible: boolean;
    variants: AppTemplateVariant[];
    versions: AppTemplateVersionSummary[];
    parameters: AppTemplateParam[];
    dependencies?: AppTemplateDependency[];
}

export interface AppTemplateImageTag {
    tag: string;
    image: string;
    class: string;
    newer: boolean;
}

export interface AppTemplateImageTagsResponse {
    repository: string;
    currentTag: string;
    truncated: boolean;
    tags: AppTemplateImageTag[];
}

export interface GetAppTemplateImageTagsParams {
    templateName: string;
    version?: string;
    variant?: string;
}

export interface CreateAppFromTemplateReq {
    projectID: string;
    projectEnv: string;
    name: string;
    template: string;
    version?: string;
    variant?: string;
    imageOverride?: string;
    params?: Record<string, unknown>;
    dependencyParams?: Record<string, Record<string, unknown>>;
}

export interface CreateAppFromTemplateResp {
    meta?: unknown;
    data: {
        app: { id: string };
        deployment: { id: string };
        dependencies?: {
            name: string;
            app: { id: string };
            deployment: { id: string };
        }[];
    };
}

export class AppTemplatesApi extends BaseApi {
    public constructor() {
        super();
    }

    /**
     * GET /app-templates/catalog
     * Fetches catalog source, revision, categories tree, and tags.
     */
    async getCatalog(signal?: AbortSignal): Promise<AppTemplateCatalog> {
        try {
            const res = await this.client.v1.get<{ meta?: unknown; data: AppTemplateCatalog }>(
                "/app-templates/catalog",
                { signal },
            );
            return res.data.data;
        } catch (error) {
            throw parseApiError(error);
        }
    }

    /**
     * GET /app-templates
     * Fetches paginated template summaries with optional filters.
     */
    async listTemplates(params: ListAppTemplatesFilter = {}, signal?: AbortSignal): Promise<ListAppTemplatesResponse> {
        try {
            const res = await this.client.v1.get<ListAppTemplatesResponse>("/app-templates", {
                params: {
                    category: params.category ?? undefined,
                    tag: params.tag ?? undefined,
                    search: params.search?.trim() ?? undefined,
                    pageOffset: params.pageOffset ?? 0,
                    pageLimit: params.pageLimit ?? 50,
                },
                signal,
            });
            const { data } = res;
            if (data.meta.page) {
                data.meta.total = data.meta.page.total;
                data.meta.pageOffset = data.meta.page.offset;
                data.meta.pageLimit = data.meta.page.limit;
                data.meta.count = data.data.length;
            }
            return data;
        } catch (error) {
            throw parseApiError(error);
        }
    }

    /**
     * GET /app-templates/{templateName}
     * Fetches detailed template information, parameters, variants, versions, and license.
     */
    async getTemplate(name: string, signal?: AbortSignal): Promise<AppTemplateDetail> {
        try {
            const res = await this.client.v1.get<{ meta?: unknown; data: AppTemplateDetail }>(
                `/app-templates/${encodeURIComponent(name)}`,
                { signal },
            );
            return res.data.data;
        } catch (error) {
            throw parseApiError(error);
        }
    }

    /**
     * GET /app-templates/{templateName}/image-tags
     * Scans Docker registry for image tags of the template's pinned repository.
     */
    async getImageTags(
        params: GetAppTemplateImageTagsParams,
        signal?: AbortSignal,
    ): Promise<AppTemplateImageTagsResponse> {
        try {
            const res = await this.client.v1.get<{ meta?: unknown; data: AppTemplateImageTagsResponse }>(
                `/app-templates/${encodeURIComponent(params.templateName)}/image-tags`,
                {
                    params: {
                        version: params.version ?? undefined,
                        variant: params.variant ?? undefined,
                    },
                    signal,
                },
            );
            return res.data.data;
        } catch (error) {
            throw parseApiError(error);
        }
    }

    /**
     * POST /projects/{projectID}/{projectEnv}/apps/from-template
     * Provisions a new app from an app template and queues its first deployment.
     */
    async createAppFromTemplate(
        req: CreateAppFromTemplateReq,
        signal?: AbortSignal,
    ): Promise<CreateAppFromTemplateResp> {
        try {
            const { projectID, projectEnv, ...body } = req;
            const res = await this.client.v1.post<CreateAppFromTemplateResp>(
                `/projects/${encodeURIComponent(projectID)}/${encodeURIComponent(projectEnv)}/apps/from-template`,
                body,
                { signal },
            );
            return res.data;
        } catch (error) {
            throw parseApiError(error);
        }
    }
}

export const appTemplatesApi = new AppTemplatesApi();
