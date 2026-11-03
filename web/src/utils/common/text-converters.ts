/**
 * Convert camelCase string to Title Case with spaces
 * e.g. "primaryTargetCompletionDate" -> "Primary Target Completion Date"
 */
export const camelCaseToTitleCase = (str: string): string => {
    return str
        .replace(/([A-Z])/g, " $1") // Add space before capital letters
        .replace(/^./, char => char.toUpperCase()) // Capitalize first letter
        .trim(); // Remove any leading/trailing spaces
};

/**
 * Convert string to kebab-case
 * e.g. "Primary Target Completion Date" -> "primary-target-completion-date"
 */
export const toKebabCase = (str: string): string => {
    return str
        .replace(/([a-z])([A-Z])/g, "$1-$2") // Add dash between lowercase and uppercase
        .replace(/[\s_]+/g, "-") // Replace spaces and underscores with dashes
        .toLowerCase();
};

/**
 * Convert kebab-case string to Title Case with spaces
 * e.g. "primary-target-completion-date" -> "Primary Target Completion Date"
 */
export const kebabCaseToTitleCase = (str: string): string => {
    return str
        .replace(/-/g, " ") // Replace dashes with spaces
        .replace(/\b\w/g, char => char.toUpperCase()) // Capitalize first letter of each word
        .trim(); // Remove any leading/trailing spaces
};

/**
 * Convert string to camelCase
 * e.g. "primary target completion date" -> "primaryTargetCompletionDate"
 */
export const toCamelCase = (str: string): string => {
    return str
        .replace(/(?:^\w|[A-Z]|\b\w)/g, (word, index) => {
            return index === 0 ? word.toLowerCase() : word.toUpperCase();
        })
        .replace(/\s+/g, "");
};

/**
 * Convert string to PascalCase
 * e.g. "primary target completion date" -> "PrimaryTargetCompletionDate"
 */
export const toPascalCase = (str: string): string => {
    return str.replace(/(?:^\w|[A-Z]|\b\w)/g, word => word.toUpperCase()).replace(/\s+/g, "");
};

/**
 * Capitalize first letter of string
 * e.g. "hello world" -> "Hello world"
 */
export const capitalize = (str: string): string => {
    return str.charAt(0).toUpperCase() + str.slice(1);
};

/**
 * Convert string to snake_case
 * e.g. "Primary Target Completion Date" -> "primary_target_completion_date"
 */
export const toSnakeCase = (str: string): string => {
    return str
        .replace(/([a-z])([A-Z])/g, "$1_$2") // Add underscore between lowercase and uppercase
        .replace(/[\s-]+/g, "_") // Replace spaces and dashes with underscores
        .toLowerCase();
};
