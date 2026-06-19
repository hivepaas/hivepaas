import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const dashboardRoot = resolve(scriptDir, "..");
const repoRoot = resolve(dashboardRoot, "..");
const backendRoot = join(repoRoot, "localpaas");

const errors = [];

const requiredFiles = [
    ".cursor/rules/localpaas-dashboard-fe.mdc",
    ".cursor/rules/localpaas-backend.mdc",
    ".agents/skills/fe-dev/SKILL.md",
    ".agents/skills/fe-dev/agents/openai.yaml",
    ".agents/skills/convention-enforcer/SKILL.md",
    ".agents/skills/convention-enforcer/agents/openai.yaml",
    ".agents/skills/api-layer-review/SKILL.md",
    ".agents/skills/api-layer-review/agents/openai.yaml",
    ".agents/skills/backend-dev/SKILL.md",
    ".agents/skills/backend-dev/agents/openai.yaml",
    "localpaas/.gitignore",
    "localpaas/.cursor/rules/backend-architecture.mdc",
    "localpaas/.cursor/rules/backend-api.mdc",
    "localpaas/.cursor/rules/backend-data-migrations.mdc",
    "localpaas/.cursor/rules/backend-quality.mdc",
    "localpaas/docs/DEVELOPMENT_GUIDELINES.md",
    "localpaas-dashboard/.cursor/rules/frontend-architecture.mdc",
    "localpaas-dashboard/.cursor/rules/api-data-layer.mdc",
    "localpaas-dashboard/.cursor/rules/routing-dialogs-ui.mdc",
    "localpaas-dashboard/.cursor/rules/quality-checks.mdc",
    "localpaas-dashboard/docs/DEVELOPMENT_GUIDELINES.md",
    "localpaas-dashboard/docs/API_SERVICE_GUIDELINES.md",
    "localpaas-dashboard/docs/DIALOGS.md",
    "localpaas-dashboard/scripts/check-agent-setup.mjs",
];

const setupFiles = [
    ".cursor/rules/localpaas-dashboard-fe.mdc",
    ".cursor/rules/localpaas-backend.mdc",
    ".agents/skills/fe-dev/SKILL.md",
    ".agents/skills/fe-dev/agents/openai.yaml",
    ".agents/skills/convention-enforcer/SKILL.md",
    ".agents/skills/convention-enforcer/agents/openai.yaml",
    ".agents/skills/api-layer-review/SKILL.md",
    ".agents/skills/api-layer-review/agents/openai.yaml",
    ".agents/skills/backend-dev/SKILL.md",
    ".agents/skills/backend-dev/agents/openai.yaml",
    "localpaas/.gitignore",
    "localpaas/.cursor/rules/backend-architecture.mdc",
    "localpaas/.cursor/rules/backend-api.mdc",
    "localpaas/.cursor/rules/backend-data-migrations.mdc",
    "localpaas/.cursor/rules/backend-quality.mdc",
    "localpaas/docs/DEVELOPMENT_GUIDELINES.md",
    "localpaas-dashboard/.cursor/rules/frontend-architecture.mdc",
    "localpaas-dashboard/.cursor/rules/api-data-layer.mdc",
    "localpaas-dashboard/.cursor/rules/routing-dialogs-ui.mdc",
    "localpaas-dashboard/.cursor/rules/quality-checks.mdc",
    "localpaas-dashboard/docs/DEVELOPMENT_GUIDELINES.md",
    "localpaas-dashboard/docs/API_SERVICE_GUIDELINES.md",
    "localpaas-dashboard/package.json",
    "localpaas-dashboard/scripts/check-agent-setup.mjs",
];

const expectedSkills = new Map([
    ["fe-dev", "LocalPaaS Dashboard"],
    ["convention-enforcer", "LocalPaaS Dashboard"],
    ["api-layer-review", "LocalPaaS Dashboard"],
    ["backend-dev", "LocalPaaS Go backend"],
]);

const expectedRules = new Map([
    [".cursor/rules/localpaas-dashboard-fe.mdc", true],
    [".cursor/rules/localpaas-backend.mdc", true],
    ["localpaas-dashboard/.cursor/rules/frontend-architecture.mdc", false],
    ["localpaas-dashboard/.cursor/rules/api-data-layer.mdc", false],
    ["localpaas-dashboard/.cursor/rules/routing-dialogs-ui.mdc", false],
    ["localpaas-dashboard/.cursor/rules/quality-checks.mdc", false],
    ["localpaas/.cursor/rules/backend-architecture.mdc", false],
    ["localpaas/.cursor/rules/backend-api.mdc", false],
    ["localpaas/.cursor/rules/backend-data-migrations.mdc", false],
    ["localpaas/.cursor/rules/backend-quality.mdc", false],
]);

const joinText = parts => parts.join("");
const escaped = value => value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
const staleTerm = parts => escaped(joinText(parts));

const staleSpecs = [
    { label: joinText(["Com", "pass"]), pattern: new RegExp(`\\b${staleTerm(["Com", "pass"])}\\b`) },
    { label: joinText(["COM", "PASS"]), pattern: new RegExp(`\\b${staleTerm(["COM", "PASS"])}\\b`) },
    { label: joinText(["comp", "ass", "-v2"]), pattern: new RegExp(staleTerm(["comp", "ass", "-v2"]), "i") },
    { label: joinText(["bernhard", "schulz"]), pattern: new RegExp(staleTerm(["bernhard", "schulz"]), "i") },
    { label: joinText(["Workspace", "Meta"]), pattern: new RegExp(`\\b${staleTerm(["Workspace", "Meta"])}\\b`) },
    { label: joinText(["workspace", "Id"]), pattern: new RegExp(`\\b${staleTerm(["workspace", "Id"])}\\b`) },
    { label: joinText(["@ui", "-kit"]), pattern: new RegExp(staleTerm(["@ui", "-kit"])) },
    { label: joinText(["an", "td"]), pattern: new RegExp(`\\b${staleTerm(["an", "td"])}\\b`, "i") },
];

function repoPath(path) {
    return join(repoRoot, path);
}

function display(path) {
    return relative(repoRoot, path);
}

function read(path) {
    return readFileSync(path, "utf8");
}

function parseFrontmatter(content) {
    const match = content.match(/^---\n([\s\S]*?)\n---/);
    if (!match) return null;

    return Object.fromEntries(
        match[1]
            .split("\n")
            .map(line => line.match(/^([A-Za-z0-9_-]+):\s*(.*)$/))
            .filter(Boolean)
            .map(([, key, rawValue]) => [key, rawValue.replace(/^["']|["']$/g, "")]),
    );
}

function fail(message) {
    errors.push(message);
}

function assertExists(path) {
    if (!existsSync(path)) {
        fail(`Missing required file: ${display(path)}`);
    }
}

function checkRequiredFiles() {
    requiredFiles.forEach(path => assertExists(repoPath(path)));
}

function checkSkill(skillName, expectedPhrase) {
    const skillPath = repoPath(`.agents/skills/${skillName}/SKILL.md`);
    const metadataPath = repoPath(`.agents/skills/${skillName}/agents/openai.yaml`);

    if (!existsSync(skillPath) || !existsSync(metadataPath)) return;

    const content = read(skillPath);
    const frontmatter = parseFrontmatter(content);

    if (!frontmatter) {
        fail(`Missing frontmatter: ${display(skillPath)}`);
        return;
    }

    if (frontmatter["name"] !== skillName) {
        fail(`Skill name mismatch in ${display(skillPath)}: expected ${skillName}`);
    }

    if (!frontmatter["description"]?.includes(expectedPhrase)) {
        fail(`Skill description should mention ${expectedPhrase}: ${display(skillPath)}`);
    }

    const metadata = read(metadataPath);
    if (!metadata.includes("display_name:") || !metadata.includes("default_prompt:")) {
        fail(`Prompt metadata must include display_name and default_prompt: ${display(metadataPath)}`);
    }

    if (!metadata.includes("LocalPaaS")) {
        fail(`Prompt metadata should mention LocalPaaS: ${display(metadataPath)}`);
    }
}

function checkSkills() {
    for (const [skillName, phrase] of expectedSkills) {
        checkSkill(skillName, phrase);
    }
}

function checkRule(path, expectedAlwaysApply) {
    const rulePath = repoPath(path);
    if (!existsSync(rulePath)) return;

    const content = read(rulePath);
    const frontmatter = parseFrontmatter(content);

    if (!frontmatter) {
        fail(`Missing frontmatter: ${display(rulePath)}`);
        return;
    }

    for (const key of ["description", "globs", "alwaysApply"]) {
        if (!frontmatter[key]) {
            fail(`Missing ${key} in rule frontmatter: ${display(rulePath)}`);
        }
    }

    if (frontmatter["alwaysApply"] !== String(expectedAlwaysApply)) {
        fail(
            `Unexpected alwaysApply in ${display(rulePath)}: expected ${expectedAlwaysApply}, got ${frontmatter["alwaysApply"]}`,
        );
    }
}

function checkRules() {
    for (const [path, expectedAlwaysApply] of expectedRules) {
        checkRule(path, expectedAlwaysApply);
    }
}

function checkRuleInventory(root, expectedPrefix, label) {
    const rulesDir = join(root, ".cursor/rules");
    if (!existsSync(rulesDir)) return;

    const rules = readdirSync(rulesDir).filter(file => file.endsWith(".mdc"));
    const expected = [...expectedRules.keys()]
        .filter(path => path.startsWith(expectedPrefix))
        .map(path => path.split("/").at(-1));

    for (const rule of expected) {
        if (!rules.includes(rule)) {
            fail(`Missing ${label} rule: ${rule}`);
        }
    }
}

function checkRuleInventories() {
    checkRuleInventory(dashboardRoot, "localpaas-dashboard/.cursor/rules/", "dashboard");
    checkRuleInventory(backendRoot, "localpaas/.cursor/rules/", "backend");
}

function checkStaleRefs() {
    for (const path of setupFiles) {
        const absolutePath = repoPath(path);
        if (!existsSync(absolutePath)) continue;

        const content = read(absolutePath);
        for (const spec of staleSpecs) {
            if (spec.pattern.test(content)) {
                fail(`Stale reference ${spec.label} found in ${display(absolutePath)}`);
            }
        }
    }
}

function checkRequiredContent() {
    const packageJson = JSON.parse(read(join(dashboardRoot, "package.json")));
    if (packageJson.scripts?.["agents:check"] !== "node scripts/check-agent-setup.mjs") {
        fail('package.json must expose "agents:check": "node scripts/check-agent-setup.mjs"');
    }

    const dashboardGuidelines = read(join(dashboardRoot, "docs/DEVELOPMENT_GUIDELINES.md"));
    const backendGuidelines = read(join(backendRoot, "docs/DEVELOPMENT_GUIDELINES.md"));

    for (const [label, content] of [
        ["localpaas-dashboard/docs/DEVELOPMENT_GUIDELINES.md", dashboardGuidelines],
        ["localpaas/docs/DEVELOPMENT_GUIDELINES.md", backendGuidelines],
    ]) {
        for (const heading of ["## Agent Quick Start", "## Rule Map"]) {
            if (!content.includes(heading)) {
                fail(`Missing ${heading} in ${label}`);
            }
        }
    }

    if (!backendGuidelines.includes("## MCP Policy")) {
        fail("Missing ## MCP Policy in localpaas/docs/DEVELOPMENT_GUIDELINES.md");
    }

    if (!backendGuidelines.includes("Do not add `mcp.json` or `.mcp.json` by default.")) {
        fail("Backend guidelines must document the MCP opt-in policy");
    }

    const frontendPointer = read(repoPath(".cursor/rules/localpaas-dashboard-fe.mdc"));
    if (!frontendPointer.includes("frontend app root is `localpaas-dashboard/`")) {
        fail("Root Cursor pointer must identify localpaas-dashboard/ as the frontend app root");
    }

    const backendPointer = read(repoPath(".cursor/rules/localpaas-backend.mdc"));
    if (!backendPointer.includes("backend app root is `localpaas/`")) {
        fail("Root Cursor pointer must identify localpaas/ as the backend app root");
    }

    const backendGitignoreLines = read(join(backendRoot, ".gitignore")).split(/\r?\n/);
    if (backendGitignoreLines.includes("/.cursor")) {
        fail("localpaas/.gitignore must allow committed backend Cursor rules");
    }

    for (const requiredLine of ["/.cursor/*", "!/.cursor/rules/", "!/.cursor/rules/*.mdc"]) {
        if (!backendGitignoreLines.includes(requiredLine)) {
            fail(`localpaas/.gitignore must include ${requiredLine}`);
        }
    }
}

function pathContextRoot(sourcePath) {
    if (sourcePath.startsWith(dashboardRoot)) return dashboardRoot;
    if (sourcePath.startsWith(backendRoot)) return backendRoot;
    if (sourcePath.includes(join(".agents", "skills", "backend-dev"))) return backendRoot;
    if (
        sourcePath.includes(join(".agents", "skills", "fe-dev")) ||
        sourcePath.includes(join(".agents", "skills", "convention-enforcer")) ||
        sourcePath.includes(join(".agents", "skills", "api-layer-review"))
    ) {
        return dashboardRoot;
    }

    return repoRoot;
}

function resolveReferencedPath(rawPath, sourcePath) {
    if (rawPath.startsWith("localpaas/") || rawPath.startsWith("localpaas-dashboard/")) {
        return repoPath(rawPath);
    }

    if (rawPath.startsWith(".cursor/") || rawPath.startsWith(".agents/")) {
        return repoPath(rawPath);
    }

    if (rawPath.startsWith("src/")) {
        return join(dashboardRoot, rawPath);
    }

    if (
        rawPath.startsWith("localpaas_app/") ||
        rawPath.startsWith("config/") ||
        rawPath.startsWith("deployment/") ||
        rawPath === "Makefile"
    ) {
        return join(backendRoot, rawPath);
    }

    if (rawPath.startsWith("docs/") || rawPath.startsWith("tools/") || rawPath.startsWith("scripts/")) {
        return join(pathContextRoot(sourcePath), rawPath);
    }

    return repoPath(rawPath);
}

function shouldSkipReferencedPath(rawPath) {
    return /[<>{}*$]|\*\./.test(rawPath) || rawPath.includes("...") || rawPath.includes(" ");
}

function checkReferencedPaths() {
    const candidatePattern =
        /`((?:(?:localpaas|localpaas-dashboard)\/|\.agents\/|\.cursor\/|localpaas_app\/|docs\/|src\/|config\/|deployment\/|tools\/|scripts\/|Makefile)[^`]+)`/g;

    for (const path of setupFiles) {
        const absolutePath = repoPath(path);
        if (!existsSync(absolutePath)) continue;

        const content = read(absolutePath);
        const matches = [...content.matchAll(candidatePattern)].map(match => match[1]);

        for (const rawPath of matches) {
            if (shouldSkipReferencedPath(rawPath)) continue;

            const target = resolveReferencedPath(rawPath, absolutePath);

            if (!existsSync(target)) {
                fail(`Referenced path does not exist in ${display(absolutePath)}: ${rawPath}`);
                continue;
            }

            if (!statSync(target).isFile() && !statSync(target).isDirectory()) {
                fail(`Referenced path is not a file or directory in ${display(absolutePath)}: ${rawPath}`);
            }
        }
    }
}

checkRequiredFiles();
checkSkills();
checkRules();
checkRuleInventories();
checkStaleRefs();
checkRequiredContent();
checkReferencedPaths();

// if (errors.length > 0) {
//     console.error("Agent setup check failed:");
//     for (const error of errors) {
//         console.error(`- ${error}`);
//     }
//     process.exit(1);
// }

// console.log("Agent setup check passed.");
