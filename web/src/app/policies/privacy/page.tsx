import React, {Suspense} from "react";
import PageWrapper from "@/app/components/clientGeneric";
import AuthArea from "@/app/components/authClient";
import {Metadata} from "next";
import {mushDbTitle} from "@/app/components/Constants";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

export const metadata: Metadata = {
    title: `Privacy Policy`,
    description: "Privacy Policy",
};

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        nextUrl: string,
    }>,
}) {

    const {nextUrl} = await params
    let err: string | undefined = undefined
    let readers: string[] = []
    try {
        readers = await GetReaderWriterNames() // Done on the server
    } catch (e) {
        err = "Failed to load page wrapper component: "+JSON.stringify(e)
    }
    return <PageWrapper props={{pageType:"policies",readers: readers}}>
        <h1>{"Privacy Policy"}</h1>
        <ErrorDisplay err={err}/>
        <p>{"Privacy policy here! Not implemented yet!"}</p>
    </PageWrapper>
}