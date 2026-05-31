import {GetReaderWriterNames} from "@/app/components/serverActions";
import {AssertSpecies} from "@/app/components/speciesClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import React from "react";
import PageWrapper from "@/app/components/clientGeneric";
import {cookies} from "next/headers";
import {SpeciesData} from "@/app/components/speciesServer";
import {ClientNewPage} from "@/app/new/[itemType]/client";
import {SessionProvider} from "@/app/components/formSubcomponents/sessionContext/session";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        itemType: string,
        species?: string[],
    }>,
}) {
    const {itemType, species} = await params
    const readers = await GetReaderWriterNames()
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
    let speciesData: SpeciesData | undefined = undefined
    if (species !== undefined){
        speciesData = await fetch(BaseExternalUrl + "/db/get/species/" + species, { // TODO: this feels way incorrect? need to un-urlencode species?
            method: 'Get',
            credentials: 'include',
            headers: {
                credentials: 'include',
                'Accept': 'application/json',
                'Cookie': allCookies, // TODO: ok?
            },
        }).then((res) => {
            if(!res.ok){
                return res.text().then(txt=>{
                    throw new Error("response not ok: "+txt)
                }).catch(err=>{
                    throw new Error("response not ok and failed to decode: ")
                })
            }
            console.log("got response")
            res.json().then((data) => {
                AssertSpecies(data)
                return data
            })
        })
    }
    return <PageWrapper props={{pageType:"new",readers: readers}}>
        <SessionProvider session={session?.value}>
            <ClientNewPage itemType={itemType} species={speciesData}/>{/*fullPage class contained within*/}
        </SessionProvider></PageWrapper>
}