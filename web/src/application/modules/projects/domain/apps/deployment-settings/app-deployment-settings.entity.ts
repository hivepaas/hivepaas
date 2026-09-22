import {
    type EAppDeploymentMethod,
    type EBuildTool,
    type EDockerfileSource,
    type ERepoType,
} from "~/projects/module-shared/enums";
import { type SettingsBaseEntity } from "~/settings/domain";

export type DeploymentDockerfile = {
    source: EDockerfileSource;
    path: string;
    content: string;
    scanPath: string;
};

export type RepoMethod = BaseDeploymentSettings & {
    activeMethod: typeof EAppDeploymentMethod.Repo;
    repoSource: {
        buildTool?: EBuildTool;
        repoType?: ERepoType;
        repoUrl: string;
        repoRef: string;
        commitHash: string;
        repoOptions: DeploymentRepoOptions;
        credentials: SettingsBaseEntity | null;
        dockerfile: DeploymentDockerfile;
        pushToRegistry: SettingsBaseEntity | null;
    };
};

export type DeploymentRepoOptions = {
    gitSubmodulesEnabled: boolean;
    gitLfsEnabled: boolean;
};

export type ImageMethod = BaseDeploymentSettings & {
    activeMethod: typeof EAppDeploymentMethod.Image;
    imageSource: {
        image: string;
        registryAuth: SettingsBaseEntity | null;
    };
};

export type BaseDeploymentSettings = {
    /**
     * What a build of this app is called: the repository it goes to, and the
     * prefix every tag of this environment carries. The rules live on the server,
     * so the form shows this rather than asking for a name.
     */
    image?: { repoName: string; tagPrefix: string } | null;

    command?: string;
    workingDir?: string;
    preDeploymentCommand?: string;
    postDeploymentCommand?: string;

    notification?: {
        successUseDefault: boolean;
        success?: { id: string; name: string } | null;
        failureUseDefault: boolean;
        failure?: { id: string; name: string } | null;
    };

    updateVer: number;
};

export type AppDeploymentSettings = RepoMethod | ImageMethod;
