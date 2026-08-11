'use client'
import {ReaderOptionsContextProvider} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import React, {ReactNode} from "react";
import {GoogleApiClient} from "@/app/components/Constants";
import TopBar from "@/app/components/TopBar";
import {QueryClient, QueryClientProvider} from "@tanstack/react-query";
import {CookiesProvider} from "react-cookie";
import {GoogleOAuthProvider} from "@react-oauth/google";
import styles from "@/app/page.module.css";
import Image from "next/image";

export function FullPage(
    {
        children
    }: {
        children: ReactNode,
    }) {
    return <>
        <div className="fullPage">
            {children}
        </div>
    </>
}


export default function PageWrapper(
    {
        props, children
    }: {
        props: {
            readers: string[]
            pageType: string
        },
        children: ReactNode,
    }) {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                staleTime: 5 * 60 * 1000, // Fresh for 5 minutes (data won't refetch in background during this time)
                gcTime: 10 * 60 * 1000,  // Stays in memory for 10 minutes after going inactive
            },
        },
    })
    return <ReaderOptionsContextProvider initialState={{options: props.readers, selected: undefined}}>
        <CookiesProvider>
            <GoogleOAuthProvider clientId={GoogleApiClient}>
                <QueryClientProvider client={queryClient}>
                    <FullPage>
                        <TopBar/>
                        {children}
                        <Footer/>
                    </FullPage>

                    {/* TODO: del? <PageTypeProvider pageType={props.pageType}>*/}

                    {/* TODO: del? </PageTypeProvider>*/}
                </QueryClientProvider>
            </GoogleOAuthProvider>
        </CookiesProvider>
    </ReaderOptionsContextProvider>
}

export function Footer() {
    return <footer className={styles.footer}>
        <a
            href="https://nextjs.org/learn?utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"
            target="_blank"
            rel="noopener noreferrer"
        >
            <Image
                aria-hidden
                src="/file.svg"
                alt="File icon"
                width={16}
                height={16}
            />
            Learn
        </a>
        <a
            href="https://vercel.com/templates?framework=next.js&utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"
            target="_blank"
            rel="noopener noreferrer"
        >
            <Image
                aria-hidden
                src="/window.svg"
                alt="Window icon"
                width={16}
                height={16}
            />
            Examples
        </a>
        <a
            href="https://nextjs.org?utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"
            target="_blank"
            rel="noopener noreferrer"
        >
            <Image
                aria-hidden
                src="/globe.svg"
                alt="Globe icon"
                width={16}
                height={16}
            />
            Go to nextjs.org →
        </a>
    </footer>
    // return <footer className={"footerActual"}>
    //     <div>{"TODO: FOOTER 1 HERE!"}</div>
    //     <div>{"TODO: FOOTER 2 HERE!"}</div>
    //     <div>{"TODO: FOOTER 3 HERE!"}</div>
    // </footer>
}