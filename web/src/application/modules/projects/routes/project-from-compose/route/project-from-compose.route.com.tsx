import { useEffect, useMemo, useRef, useState } from "react";

import { Button, Checkbox, Label } from "@components/ui";
import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { OctagonXIcon } from "lucide-react";
import { useNavigate, useParams } from "react-router";
import invariant from "tiny-invariant";
import { specImportErrorCode } from "~/operations/api/hooks";
import { ComposeImportCommands } from "~/operations/data";
import type {
    ComposeFileInput,
    ComposeImportBody,
    ComposeImportResult,
    ComposeImportReview,
    ComposeServiceInput,
    ComposeVariableInput,
    SpecImportSelection,
} from "~/operations/domain";
import { SpecImportPlanTree, buildImportTree, leafPaths, selectionOf } from "~/operations/routes/export";
import { ProjectsQueries } from "~/projects/data/queries";
import { type ProjectEnvEntity } from "~/projects/domain";
import { useProjectEnvFilter } from "~/projects/module-shared/hooks";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { ValidationProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "@infrastructure/exceptions/http";

import {
    ComposeFiles,
    ComposeResult,
    ComposeServices,
    ComposeSource,
    ComposeTarget,
    type ComposeTargetValue,
    ComposeVariables,
} from "../building-blocks";

/** How long the review waits for typing to stop before the file is read again. */
const REVALIDATE_DELAY_MS = 500;

const EVERYTHING: SpecImportSelection = { include: [], exclude: [] };

const EMPTY_ENVS: ProjectEnvEntity[] = [];

/** A new env's color until another is picked. */
const DEFAULT_ENV_COLOR = "#64748b";

/** The env offered at first: the header's, when one is chosen there, else the project's first. */
function defaultTarget(envs: ProjectEnvEntity[], headerEnv: string): ComposeTargetValue {
    const env = envs.find(item => item.name === headerEnv) ?? envs[0];

    return env
        ? { env: env.name, newEnv: false, color: DEFAULT_ENV_COLOR }
        : { env: "", newEnv: true, color: DEFAULT_ENV_COLOR };
}

/** What a refusal of validate says, by where it belongs on the page. */
interface ReadErrors {
    name?: string;
    env?: string;
    /** About the file itself, or a field shown nowhere else. */
    file?: string;
}

/**
 * Splits a refusal by field: the project's name and env are said under their
 * inputs, in words of this page; anything else under the file, field by field.
 */
function readErrorsOf(error: Error | undefined): ReadErrors {
    if (!error) {
        return {};
    }
    if (!(error instanceof HttpException) || !(error.problem instanceof ValidationProblemApiResponse)) {
        return { file: error.message };
    }
    const out: ReadErrors = {};
    const rest: string[] = [];
    for (const item of error.problem.errors) {
        if (item.path === "project.name") {
            out.name =
                item.code === "ERR_VLD_VALUE_REQUIRED"
                    ? "Name the project: the compose file has no name of its own (name:)."
                    : item.message;
        } else if (item.path === "project.env") {
            out.env = item.message;
        } else {
            rest.push(`${item.path}: ${item.message}`);
        }
    }
    if (rest.length > 0) {
        out.file = rest.join("; ");
    }

    return out;
}

/** A key of what a body asks, File objects named by what tells them apart. */
function keyOf(body: ComposeImportBody): string {
    const files = Object.entries(body.files).map(([path, file]) =>
        typeof file === "string" ? [path, file] : [path, file.name, file.size, file.lastModified],
    );

    return JSON.stringify({ ...body, files });
}

/** A new project from a Docker Compose file. */
export function ProjectFromComposeRoute() {
    return <ComposeImportPage />;
}

/** A Docker Compose file's services added to an env of the route's project. */
export function ProjectAppsFromComposeRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    return (
        <ComposeImportPage
            key={projectId}
            projectId={projectId}
        />
    );
}

/**
 * Paste a compose file, see what each service becomes and what the file needs,
 * and create the apps - in a new project's one env, or in an env of the project
 * given - once the plan is what was meant. Nothing is written before.
 */
function ComposeImportPage({ projectId }: { projectId?: string }) {
    const intoProject = projectId !== undefined;
    const navigate = useNavigate();
    const { data: projectData } = ProjectsQueries.useFindOneById(
        { projectID: projectId ?? "" },
        { enabled: intoProject },
    );
    const project = intoProject ? projectData?.data : undefined;
    const projectEnvs = project?.envs ?? EMPTY_ENVS;
    const { selectedEnv: headerEnv, setSelectedEnv } = useProjectEnvFilter(projectId ?? "");
    const [targetInput, setTarget] = useState<ComposeTargetValue | undefined>();
    const target = targetInput ?? defaultTarget(projectEnvs, headerEnv);

    const [compose, setCompose] = useState("");
    const [dotEnv, setDotEnv] = useState("");
    const [projectName, setProjectName] = useState("");
    const [envName, setEnvName] = useState("production");
    const [profiles, setProfiles] = useState<string[]>([]);
    const [variables, setVariables] = useState<Record<string, ComposeVariableInput>>({});
    const [files, setFiles] = useState<Record<string, ComposeFileInput>>({});
    const [services, setServices] = useState<Record<string, ComposeServiceInput>>({});
    const [deploy, setDeploy] = useState(true);
    const [checked, setChecked] = useState<Set<string> | undefined>();
    const [review, setReview] = useState<ComposeImportReview | undefined>();
    const [reviewKey, setReviewKey] = useState<string | undefined>();
    const [validateError, setValidateError] = useState<Error | undefined>();
    const [planChanged, setPlanChanged] = useState(false);
    const [result, setResult] = useState<ComposeImportResult | undefined>();
    const [revision, setRevision] = useState(0);

    const plan = review?.plan;
    const roots = useMemo(() => (plan ? buildImportTree(plan.nodes) : []), [plan]);
    const selection = useMemo(() => (checked ? selectionOf(roots, checked) : EVERYTHING), [roots, checked]);

    const body = useMemo<ComposeImportBody>(
        () => ({
            projectId,
            compose,
            dotEnv,
            files,
            variables,
            project: intoProject
                ? {
                      name: "",
                      env: target.env.trim(),
                      newEnv: target.newEnv,
                      envColor: target.newEnv ? target.color : undefined,
                  }
                : { name: projectName.trim(), env: envName.trim() },
            profiles,
            services,
            selection: selection ?? EVERYTHING,
            deploy,
        }),
        [
            projectId,
            intoProject,
            target,
            compose,
            dotEnv,
            files,
            variables,
            projectName,
            envName,
            profiles,
            services,
            selection,
            deploy,
        ],
    );
    const requestKey = useMemo(() => keyOf(body), [body]);

    const { mutateAsync: validate, isPending: isValidating } = ComposeImportCommands.useValidateCompose();
    const { mutate: apply, isPending: isApplying } = ComposeImportCommands.useApplyCompose({
        onSuccess: response => {
            const failed = response.data.plan.nodes.some(node => node.outcome === "failed");
            if (projectId && !failed && !response.data.warning) {
                // Added cleanly: the apps, in the env they went into.
                setSelectedEnv(target.env.trim());
                void navigate(ROUTE.projects.single.apps.$route(projectId));

                return;
            }
            setResult(response.data);
        },
        onError: error => {
            if (specImportErrorCode(error) === "ERR_SPEC_IMPORT_PLAN_CHANGED") {
                setPlanChanged(true);
                setReviewKey(undefined);
                setRevision(value => value + 1);
            }
        },
    });

    // The file is read again whenever what it is read with changes, once the
    // typing stops. Only the latest answer is kept.
    const latest = useRef(0);
    // The plan's leaves seen so far: one that appears - a service added to the
    // file - is checked, as everything is at first; one unchecked stays so.
    const seenLeaves = useRef(new Set<string>());
    useEffect(() => {
        if (compose.trim() === "" || result || (intoProject && body.project.env === "")) {
            return;
        }
        const requestID = ++latest.current;
        const key = requestKey;
        const timer = window.setTimeout(() => {
            validate(body)
                .then(response => {
                    if (requestID !== latest.current) {
                        return;
                    }
                    setReview(response.data);
                    setReviewKey(key);
                    setValidateError(undefined);
                    const nodes = response.data.plan?.nodes;
                    if (nodes) {
                        const leaves = leafPaths(buildImportTree(nodes));
                        const seen = seenLeaves.current;
                        seenLeaves.current = new Set(leaves);
                        setChecked(previous => {
                            const kept = leaves.filter(path => !seen.has(path) || previous?.has(path));

                            return new Set(kept);
                        });
                    }
                })
                .catch((error: unknown) => {
                    if (requestID !== latest.current) {
                        return;
                    }
                    setValidateError(error instanceof Error ? error : new Error(String(error)));
                });
        }, REVALIDATE_DELAY_MS);

        return () => {
            window.clearTimeout(timer);
        };
        // body is what requestKey names; asking again on each of its renders is not wanted.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [requestKey, revision, result, validate]);

    const startOver = () => {
        latest.current++;
        setCompose("");
        setDotEnv("");
        setProjectName("");
        setTarget(undefined);
        setProfiles([]);
        setVariables({});
        setFiles({});
        setServices({});
        setChecked(undefined);
        seenLeaves.current = new Set();
        setReview(undefined);
        setReviewKey(undefined);
        setValidateError(undefined);
        setPlanChanged(false);
        setResult(undefined);
    };

    const readErrors = readErrorsOf(validateError);
    const summary = plan?.summary ?? {};
    const blocked = summary["blocked"] ?? 0;
    const accepted = (summary["skipped"] ?? 0) + (summary["fixable"] ?? 0) + (summary["warning"] ?? 0);
    const isCurrent = reviewKey === requestKey && !isValidating;
    const missingVariables = review?.variables.filter(
        variable => variable.required && !variable.given && !variables[variable.name]?.value,
    );
    const canCreate = plan !== undefined && isCurrent && blocked === 0 && selection !== undefined && !isApplying;

    if (result) {
        return (
            <div className={cn(listBox, "flex flex-col gap-6")}>
                <ComposeResult
                    result={result}
                    intoProject={intoProject}
                    onStartOver={startOver}
                />
            </div>
        );
    }

    return (
        <div className={cn(listBox, "flex flex-col gap-6")}>
            <div className="rounded-lg border bg-background p-4">
                <div className="flex flex-col items-start gap-6">
                    <div>
                        <p className="text-base font-medium text-foreground">
                            {intoProject ? "New apps from Docker Compose" : "New project from Docker Compose"}
                        </p>
                        <p className="mt-1 max-w-[900px] text-sm text-muted-foreground">
                            {intoProject
                                ? "Paste a compose file. Each service becomes an app of the env you choose, reached by the same name; nothing the project has is changed."
                                : "Paste a compose file. Each service becomes an app of the project's one env, reached by the same name."}{" "}
                            Nothing is created until you confirm: first you see what each service becomes, what the file
                            needs, and what HivePaaS cannot carry over.
                        </p>
                    </div>

                    <ComposeSource
                        compose={compose}
                        onComposeChange={setCompose}
                        dotEnv={dotEnv}
                        onDotEnvChange={setDotEnv}
                        projectName={projectName}
                        onProjectNameChange={setProjectName}
                        fileName={review?.project.fileName ?? ""}
                        envName={envName}
                        onEnvNameChange={setEnvName}
                        profiles={review?.profiles ?? []}
                        selectedProfiles={profiles}
                        onProfilesChange={setProfiles}
                        error={readErrors.file}
                        nameError={readErrors.name}
                        envError={readErrors.env}
                        target={
                            intoProject ? (
                                <ComposeTarget
                                    projectName={project?.name ?? ""}
                                    envs={projectEnvs}
                                    value={target}
                                    onChange={setTarget}
                                    envError={readErrors.env}
                                />
                            ) : undefined
                        }
                        isReading={isValidating}
                    />

                    {review && (
                        <>
                            <ComposeVariables
                                variables={review.variables}
                                inputs={variables}
                                onChange={setVariables}
                            />
                            <ComposeFiles
                                needs={review.needs}
                                files={files}
                                onChange={setFiles}
                            />
                            <ComposeServices
                                services={review.services}
                                inputs={services}
                                onChange={setServices}
                            />
                        </>
                    )}

                    {review && !plan && missingVariables && missingVariables.length > 0 && (
                        <p className="text-sm text-muted-foreground">
                            Give {missingVariables.map(variable => variable.name).join(", ")} a value to see the plan.
                        </p>
                    )}

                    {review && plan && checked && (
                        <div className="flex w-full flex-col gap-2">
                            <div className="flex flex-wrap items-baseline justify-between gap-2">
                                <p className="text-sm font-medium text-foreground">
                                    Project <code className="font-mono">{review.project.name}</code>, env{" "}
                                    <code className="font-mono">{review.project.env}</code>
                                    {intoProject && review.project.newEnv ? " (new)" : ""}
                                </p>
                                <p className="text-xs text-muted-foreground">{isValidating ? "Checking…" : ""}</p>
                            </div>
                            <SpecImportPlanTree
                                roots={roots}
                                checked={checked}
                                onChange={setChecked}
                            />
                        </div>
                    )}

                    {plan && (
                        <div className="flex w-full flex-col items-start gap-3">
                            <div className="flex items-start gap-2">
                                <Checkbox
                                    id="compose-deploy"
                                    className="mt-0.5"
                                    checked={deploy}
                                    onCheckedChange={value => {
                                        setDeploy(value === true);
                                    }}
                                />
                                <div className="flex flex-col">
                                    <Label
                                        htmlFor="compose-deploy"
                                        className="text-sm font-normal"
                                    >
                                        Deploy the apps once they are created
                                    </Label>
                                    <p className="text-xs text-muted-foreground">
                                        Otherwise they start on a placeholder image, as an app created by hand does.
                                    </p>
                                </div>
                            </div>

                            {planChanged && (
                                <p className="w-full rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400">
                                    Something changed on this installation since the plan was made, and so did the plan.
                                    Review it, and confirm again.
                                </p>
                            )}
                            {blocked > 0 && (
                                <p className="flex items-center gap-2 text-sm text-destructive">
                                    <OctagonXIcon className="size-4" />
                                    {blocked} {blocked === 1 ? "issue blocks" : "issues block"}{" "}
                                    {intoProject ? "adding the apps" : "creating the project"}.
                                </p>
                            )}
                            {blocked === 0 && accepted > 0 && (
                                <p className="max-w-[720px] text-sm text-muted-foreground">
                                    Creating it accepts {accepted} {accepted === 1 ? "issue" : "issues"}: what is left
                                    out is not created, what is cleared is created without it, and a warning changes
                                    nothing but may not be what the file meant.
                                </p>
                            )}

                            <PermissionTooltipAction
                                id={MODULE_IDS.Project}
                                action="write"
                            >
                                {({ isDenied }) => (
                                    <Button
                                        type="button"
                                        className="min-w-[120px]"
                                        disabled={!canCreate || isDenied}
                                        isLoading={isApplying}
                                        onClick={() => {
                                            if (isDenied) {
                                                return;
                                            }
                                            setPlanChanged(false);
                                            apply({ ...body, planHash: plan.planHash, acceptIssues: accepted > 0 });
                                        }}
                                    >
                                        {accepted > 0 && blocked === 0
                                            ? `${intoProject ? "Add" : "Create"} and accept ${accepted} ${accepted === 1 ? "issue" : "issues"}`
                                            : intoProject
                                              ? "Add apps"
                                              : "Create project"}
                                    </Button>
                                )}
                            </PermissionTooltipAction>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
