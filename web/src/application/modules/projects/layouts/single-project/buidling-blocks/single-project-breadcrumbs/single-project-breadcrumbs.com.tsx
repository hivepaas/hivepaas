import { memo } from "react";

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@components/ui";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

function View({ project }: Props) {
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.projects.list.$route}>Projects</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbPage>{project.name === "" ? "<unset>" : project.name}</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
    );
}

interface Props {
    project: {
        name: string;
    };
}

export const SingleProjectBreadcrumbs = memo(View);
