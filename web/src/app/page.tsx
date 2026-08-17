import Image from "next/image";
import styles from "./page.module.css";
import PageWrapper, {Footer} from "@/app/components/clientGeneric";
import React from "react";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import {cookies} from "next/headers";
import {BaseExternalUrl} from "@/app/components/Constants";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

// TODO: use <Link> to link btwn pages
export default async function Page(){
  //export default async function Home() {
  const cookieStore = await cookies()
  const session = cookieStore.get('_gothic_session')
  const readers = await GetReaderWriterNames() // TODO; ensure works
  return <PageWrapper props={{pageType: "view", readers: readers}}>
      <div className={styles.page}>
        <main className={styles.main}>
          {/* TODO: if not logged in, redirect to login page!*/}
          {/* TODO: login page link*/}
          <Image
              className={styles.logo/* TODO: remove or replace image*/}
              src="/next.svg"
              alt="Next.js logo"
              width={180}
              height={38}
          />
          <ol>
            <li>
              Welcome to the homepage for my fungi cultivation tracking website. {/*Get started by editing <code>src/app/page.tsx</code>.*/}
            </li>
            {/* TODO: next line is not working properly!!!!*/}
            <li>{(session===undefined || session.value==="") ? <a href={BaseExternalUrl+"/login"}>{"Login"}</a>:
                <>
                  <span>{"Logged in. "}</span>
                  <a href={BaseExternalUrl+"/login"}>{"Login again"}</a>
                </>}
            </li>
          </ol>

          {/*TODO: remove<div className={styles.ctas}>*/}
          {/*  <a*/}
          {/*      className={styles.primary}*/}
          {/*      href="https://vercel.com/new?utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"*/}
          {/*      target="_blank"*/}
          {/*      rel="noopener noreferrer"*/}
          {/*  >*/}
          {/*    <Image*/}
          {/*        className={styles.logo}*/}
          {/*        src="/vercel.svg"*/}
          {/*        alt="Vercel logomark"*/}
          {/*        width={20}*/}
          {/*        height={20}*/}
          {/*    />*/}
          {/*    Deploy now*/}
          {/*  </a>*/}
          {/*  <a*/}
          {/*      href="https://nextjs.org/docs?utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"*/}
          {/*      target="_blank"*/}
          {/*      rel="noopener noreferrer"*/}
          {/*      className={styles.secondary}*/}
          {/*  >*/}
          {/*    Read our docs*/}
          {/*  </a>*/}
        </main>
      </div>
  </PageWrapper>
}
