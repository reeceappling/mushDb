import React, {Suspense} from "react";
import PageWrapper from "@/app/components/clientGeneric";
import AuthArea from "@/app/components/authClient";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        nextUrl: string,
    }>,
}) {
    const {nextUrl} = await params
    return <PageWrapper props={{pageType:"auth",readers: []}}>{/* TODO: remove readers? */}
        <Suspense fallback={<p>{"Loading..."}</p>}>
                <AuthArea successUrl={nextUrl} loggedIn={false}/>
        </Suspense>
    </PageWrapper>
}