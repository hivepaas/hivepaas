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
    /** Deploying this grants the app kernel capabilities, sysctls, ulimits or
     *  the GPU, which needs Write on the Cluster module. */
    requiresCapabilities?: boolean;
    /** Deploying this lets an app start containers through HivePaaS's Docker API
     *  proxy, which also needs Write on the Cluster module. */
    requiresDockerApi?: boolean;
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
    /** Set on an "app" parameter: the engine the apps offered for it must run. */
    engine?: string;
    generated: boolean;
    options?: AppTemplateParamOption[];
}

/** An app of the environment, as a parameter that names one offers it. */
export interface EnvApp {
    id: string;
    name: string;
    key: string;
    env: string;
    /** What the app runs, absent for an app created without a template. */
    engine?: string;
    /** Apps created to serve this one - a template's dependencies. */
    logicalChildApps?: EnvApp[];
    /** Apps that name this one as their parent - a preview deployment. */
    childApps?: EnvApp[];
}

export interface ListEnvAppsParams {
    projectID: string;
    projectEnv: string;
    search?: string;
    pageLimit?: number;
}

export interface AppTemplateUlimit {
    name: string;
    soft: number;
    hard: number;
}

/** What deploying a template grants the app beyond what a container ordinarily
 *  gets. A version cannot change it, so this is what will be granted. */
export interface AppTemplateCapabilities {
    /** Named as docker names them, without the CAP_ prefix: NET_ADMIN. */
    capabilityAdd?: string[];
    capabilityDrop?: string[];
    sysctls?: Record<string, string>;
    ulimits?: AppTemplateUlimit[];
    enableGPU?: boolean;
    oomScoreAdj?: number;
}

/** What a template lets its app do through HivePaaS's Docker API proxy: start
 *  containers of these images, sharing these directories, within these limits.
 *  The app never gets the Docker socket itself. Like capabilities, a version
 *  cannot change it. */
export interface AppTemplateDockerApi {
    /** Patterns over the images children may run; "*" is any image. */
    images: string[];
    /** Directories of the app a child may bind. */
    sharedDirs?: string[];
    /** "env": children may also join the app's env network. */
    networks?: string[];
    /** Groups of endpoints beyond the core: exec, files, volumes, networks, nestedSocket. */
    allow?: string[];
    /** Absent where the template leaves HivePaaS's default. */
    containers?: number;
    memory?: string;
    cpus?: number;
}

/** One address a template claims on the cluster itself, beside the web
 *  addresses the reverse proxy serves. Two apps cannot share one. */
export interface AppTemplatePort {
    /** Inside the container, and on the nodes. */
    target: number;
    published: number;
    /** The parameter that chooses the published port, when one does: `published`
     *  is then only its default, and what the person types is what is claimed. */
    publishedParam?: string;
    protocol: string;
    /** "ingress" answers on every node, "host" only on the node running the app. */
    publishMode: string;
}

export interface AppTemplateDependency {
    name: string;
    title: string;
    template: string;
    templateTitle?: string;
    version?: string;
    variant?: string;
    parameters?: AppTemplateParam[];
    /** What this dependency's app is granted: the same request creates it. */
    capabilities?: AppTemplateCapabilities | null;
    dockerApi?: AppTemplateDockerApi | null;
    /** Ports this dependency's app claims on the cluster. */
    publishedPorts?: AppTemplatePort[] | null;
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
    /** Null for the templates that ask for nothing, which is most of them. */
    capabilities?: AppTemplateCapabilities | null;
    /** Null for the templates whose app starts no containers of its own. */
    dockerApi?: AppTemplateDockerApi | null;
    /** Empty for the templates that publish nothing, which is most of them. */
    publishedPorts?: AppTemplatePort[] | null;
}

export interface AppTemplateImageTag {
    /** What goes back as imageTag when the app is created. */
    tag: string;
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
    /** A tag, not a reference: the repository is the template's. */
    imageTag?: string;
    params?: Record<string, unknown>;
    dependencyParams?: Record<string, Record<string, unknown>>;
    /**
     * Deletes what a previous install of these apps left in their directories
     * before the new ones are created. It is how the preflight findings are
     * answered, and it is off unless asked for.
     */
    resetStorage?: boolean;
}

/** One app of a request whose directory in a volume already holds something. */
export interface PreflightStorageFinding {
    app: string;
    appKey: string;
    /**
     * A database started on somebody else's data keeps the password that data was
     * created with, and the one being generated now will not open it.
     */
    isDatabase: boolean;
    volume: { id: string; name: string };
    path: string;
}

/** A refusal the creation would raise, reported instead of raised. */
export interface PreflightIssue {
    code: string;
    detail: string;
}

export interface PreflightAppFromTemplateResp {
    meta?: unknown;
    data: {
        storage: PreflightStorageFinding[];
        storageUnchecked?: PreflightStorageFinding[];
        issues?: PreflightIssue[];
    };
}

/** What the preflight saw, and what it could not see. */
export interface PreflightResult {
    storage: PreflightStorageFinding[];
    /**
     * What the creation would refuse: a domain already served, a port already
     * held, permission the caller does not have. Reported together, because a
     * dialog that says "and also" three times is three trips through the form.
     */
    issues: PreflightIssue[];
    /**
     * Storage nothing could be seen of, usually a node that could not be
     * reached. Not the same as there being nothing there, which is why it is
     * reported rather than dropped.
     */
    unchecked: PreflightStorageFinding[];
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
                    pageLimit: params.pageLimit ?? undefined,
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

    /**
     * POST /projects/{projectID}/{projectEnv}/apps/from-template/preflight
     * What creating this request would run into, without creating anything: which
     * of its apps would be given a directory that already holds data.
     */
    async preflightAppFromTemplate(req: CreateAppFromTemplateReq, signal?: AbortSignal): Promise<PreflightResult> {
        try {
            const { projectID, projectEnv, ...body } = req;
            const res = await this.client.v1.post<PreflightAppFromTemplateResp>(
                `/projects/${encodeURIComponent(projectID)}/${encodeURIComponent(projectEnv)}/apps/from-template/preflight`,
                body,
                { signal },
            );
            return {
                storage: res.data.data.storage,
                unchecked: res.data.data.storageUnchecked ?? [],
                issues: res.data.data.issues ?? [],
            };
        } catch (error) {
            throw parseApiError(error);
        }
    }

    /**
     * GET /projects/{projectID}/{projectEnv}/apps
     * The environment's apps, for a parameter that names one. The full listing
     * rather than the base one, because only this carries `engine` - which is what
     * the list is narrowed by. Searching is the server's, so a long list stays one
     * page.
     *
     * Children are asked for and flattened: a database created as the dependency of
     * another app is exactly the kind of thing a replica is pointed at, and the
     * listing nests those rather than returning them on their own.
     */
    async listEnvApps(params: ListEnvAppsParams, signal?: AbortSignal): Promise<EnvApp[]> {
        try {
            const res = await this.client.v1.get<{ meta?: unknown; data: EnvApp[] }>(
                `/projects/${encodeURIComponent(params.projectID)}/${encodeURIComponent(params.projectEnv)}/apps`,
                {
                    params: {
                        search: params.search ?? undefined,
                        pageLimit: params.pageLimit ?? undefined,
                    },
                    signal,
                },
            );
            return res.data.data;
        } catch (error) {
            throw parseApiError(error);
        }
    }
}

export const appTemplatesApi = new AppTemplatesApi();
