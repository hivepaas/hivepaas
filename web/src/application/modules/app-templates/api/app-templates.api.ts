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

export interface ListAppTemplatesMeta {
    total: number;
    count: number;
    pageOffset: number;
    pageLimit: number;
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
            return res.data;
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
}

export const appTemplatesApi = new AppTemplatesApi();
