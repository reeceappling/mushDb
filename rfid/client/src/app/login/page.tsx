import React, {Suspense} from "react";
import PageWrapper from "@/app/components/clientGeneric";
import AuthArea from "@/app/components/authClient";
// import {CookiesProvider} from "react-cookie";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        nextUrl: string,
    }>,
}) {
    // TODO: TOP BAR?
    const {nextUrl} = await params
    return <PageWrapper props={{pageType:"login",readers: []}}>{/* TODO: remove readers? */}
        <Suspense fallback={<p>{"Loading..."}</p>}>
            <div>{"HOMEPAGE STUFF HERE: TODO: THIS"}</div>
            <AuthArea successUrl={nextUrl} loggedIn={false}/>
        </Suspense>
    </PageWrapper>
}