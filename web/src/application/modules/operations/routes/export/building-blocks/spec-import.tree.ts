import type { SpecImportNode, SpecImportSelection } from "~/operations/domain";

/**
 * The plan's nodes as the tree the operator picks from: global; each project,
 * its own settings and its envs; each env, its own settings and its apps.
 *
 * A node's parent is the nearest node whose path is a prefix of its own, which
 * is how the server lays paths out - `projects/a/envs/dev` is below
 * `projects/a`, `projects/a/envs/dev/apps/web` below that.
 */
export interface SpecImportTreeNode {
    node: SpecImportNode;
    depth: number;
    children: SpecImportTreeNode[];
}

export type SpecImportCheckState = "checked" | "unchecked" | "indeterminate";

export function buildImportTree(nodes: SpecImportNode[]): SpecImportTreeNode[] {
    const byPath = new Map<string, SpecImportTreeNode>();
    const roots: SpecImportTreeNode[] = [];

    for (const node of nodes) {
        byPath.set(node.path, { node, depth: 0, children: [] });
    }
    for (const node of nodes) {
        const treeNode = byPath.get(node.path);
        if (!treeNode) {
            continue;
        }
        const parent = parentOf(node.path, byPath);
        if (parent) {
            parent.children.push(treeNode);
        } else {
            roots.push(treeNode);
        }
    }
    setDepth(roots, 0);

    return roots;
}

function parentOf(path: string, byPath: Map<string, SpecImportTreeNode>): SpecImportTreeNode | undefined {
    let end = path.lastIndexOf("/");
    while (end > 0) {
        const parent = byPath.get(path.slice(0, end));
        if (parent) {
            return parent;
        }
        end = path.lastIndexOf("/", end - 1);
    }

    return undefined;
}

function setDepth(nodes: SpecImportTreeNode[], depth: number) {
    for (const treeNode of nodes) {
        treeNode.depth = depth;
        setDepth(treeNode.children, depth + 1);
    }
}

/** Every node, parents before their children, as the tree is drawn. */
export function flattenImportTree(roots: SpecImportTreeNode[]): SpecImportTreeNode[] {
    return roots.flatMap(root => [root, ...flattenImportTree(root.children)]);
}

/**
 * The nodes a checkbox is kept for: those with nothing below them. A parent is
 * checked when everything below it is, which is also when its own record is
 * taken - a project's name, an env's colour.
 */
export function leafPaths(roots: SpecImportTreeNode[]): string[] {
    return roots.flatMap(root => (root.children.length === 0 ? [root.node.path] : leafPaths(root.children)));
}

export function checkStateOf(treeNode: SpecImportTreeNode, checked: ReadonlySet<string>): SpecImportCheckState {
    const leaves = leafPaths([treeNode]);
    const count = leaves.filter(path => checked.has(path)).length;
    if (count === 0) {
        return "unchecked";
    }

    return count === leaves.length ? "checked" : "indeterminate";
}

/** Checks or unchecks a node and everything below it. */
export function toggleNode(treeNode: SpecImportTreeNode, checked: ReadonlySet<string>, value: boolean): Set<string> {
    const next = new Set(checked);
    for (const path of leafPaths([treeNode])) {
        if (value) {
            next.add(path);
        } else {
            next.delete(path);
        }
    }

    return next;
}

/**
 * The selection the checkboxes stand for, or undefined when nothing is checked.
 *
 * Only include is used. What is left unchecked is not excluded, just not
 * asked for - so the import can still pull in a setting from it that something
 * checked refers to and this installation lacks, such as the certificate of an
 * app. A fully checked subtree is named once, by its root.
 */
export function selectionOf(
    roots: SpecImportTreeNode[],
    checked: ReadonlySet<string>,
): SpecImportSelection | undefined {
    const leaves = leafPaths(roots);
    if (leaves.every(path => checked.has(path))) {
        return { include: [], exclude: [] };
    }

    const include: string[] = [];
    const visit = (treeNode: SpecImportTreeNode) => {
        const state = checkStateOf(treeNode, checked);
        if (state === "checked") {
            include.push(treeNode.node.path);
        } else if (state === "indeterminate") {
            treeNode.children.forEach(visit);
        }
    };
    roots.forEach(visit);

    return include.length > 0 ? { include, exclude: [] } : undefined;
}
