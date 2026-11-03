const HTTP_URL_PATTERN = /^https?:\/\//i;

export function toCommandTemplateExternalHref(link: string): string {
    const trimmed = link.trim();

    if (!trimmed) {
        return "";
    }

    if (HTTP_URL_PATTERN.test(trimmed)) {
        return trimmed;
    }

    return `https://${trimmed}`;
}
