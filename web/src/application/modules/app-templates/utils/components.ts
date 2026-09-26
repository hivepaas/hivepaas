import type { AppTemplateComponent, AppTemplateDetail } from "../api";

/**
 * How one component is named where what it is granted is listed: the template,
 * then the component, since "Worker" alone does not say whose.
 */
export function componentLabel(template: AppTemplateDetail, component: AppTemplateComponent): string {
    return `${template.title} ${component.title}`;
}
