/**
 * The settings an environment lists that its apps can use: those it inherits
 * from the project or the installation, and its own made inheritable. Its own
 * that are not inheritable are listed too, and an app saved with one of them is
 * refused, the setting "not found".
 */
export function appUsableSettings<T extends { inheritable?: boolean }>(settings: T[]): T[] {
    return settings.filter(setting => setting.inheritable === true);
}
