import { memo } from "react";

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@components/ui";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

function View({ app, project }: Props) {
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.projects.list.$route}>Projects</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.projects.single.apps.$route(project.id)}>{project.name}</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbPage>{app.name === "" ? "<unset>" : app.name}</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
    );
}

interface Props {
    app: {
        name: string;
    };
    project: {
        id: string;
        name: string;
    };
}

export const SingleAppBreadcrumbs = memo(View);
