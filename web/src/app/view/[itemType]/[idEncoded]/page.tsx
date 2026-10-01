import React from "react";
import {BaseExternalUrl} from "@/app/components/Constants";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import {cookies} from 'next/headers'
import {MainViewArea} from "@/app/view/[itemType]/[idEncoded]/client";
import {CookiesProvider} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {Metadata} from "next";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {WaterJarData} from "@/app/components/waterJarServer";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {BagData} from "@/app/components/bagServer";
import {FruitData} from "@/app/components/fruitServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {JarData} from "@/app/components/jarServer";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {LcData} from "@/app/components/lcServer";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {LcSyringeData} from "@/app/components/lcSyringeServer";
import {MssData} from "@/app/components/mssServer";
import {PcRunData} from "@/app/components/pcRunServer";
import {PlateData} from "@/app/components/plateServer";
import {PlugsData} from "@/app/components/plugsServer";
import {ProjectData} from "@/app/components/projectServer";
import {SaleData} from "@/app/components/saleServer";
import {SlantData} from "@/app/components/slantServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {SporeSwabData} from "@/app/components/sporeSwabServer";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {TransferData} from "@/app/components/transferServer";
import {UserData} from "@/app/components/userServer";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

type Props = {
    params: Promise<{
        itemType: string
        idEncoded: string // urlEncoded
    }>
};
// Next.js runs this first to set the tab title
export async function generateMetadata({ params }: Props): Promise<Metadata> {
    const {itemType, idEncoded} = await params
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
    const decoded = decodeURI(idEncoded)
    const {doIndex, desc} = await getData(itemType, idEncoded, allCookies).then(resultData=>{
        const dind = resultData.doIndex
        switch(itemType){
            case 'agarBatch':
                return {doIndex:dind,desc:(new AgarBatchData(resultData.data)).description()}
            case 'agarRecipe':
                return {doIndex:dind,desc:(new AgarRecipeData(resultData.data)).description()}
            case 'bag':
                return {doIndex:dind,desc:(new BagData(resultData.data)).description()}
            case 'fruit':
                return {doIndex:dind,desc:(new FruitData(resultData.data)).description()}
            case 'fruitingChamber':
                return {doIndex:dind,desc:(new FruitingChamberData(resultData.data)).description()}
            case 'grainBatch':
                return {doIndex:dind,desc:(new GrainBatchData(resultData.data)).description()}
            case 'jar':
                return {doIndex:dind,desc:(new JarData(resultData.data)).description()}
            case 'jarRecipe':
                return {doIndex:dind,desc:(new JarRecipeData(resultData.data)).description()}
            case 'lc':
                return {doIndex:dind,desc:(new LcData(resultData.data)).description()}
            case 'lcRecipe':
                return {doIndex:dind,desc:(new LcRecipeData(resultData.data)).description()}
            case 'lcSyringe':
                return {doIndex:dind,desc:(new LcSyringeData(resultData.data)).description()}
            case 'mss':
                return {doIndex:dind,desc:(new MssData(resultData.data)).description()}
            case 'pcRun':
                return {doIndex:dind,desc:(new PcRunData(resultData.data)).description()}
            case 'plate':
                return {doIndex:dind,desc:(new PlateData(resultData.data)).description()}
            case 'plugs':
                return {doIndex:dind,desc:(new PlugsData(resultData.data)).description()}
            case 'project':
                return {doIndex:dind,desc:(new ProjectData(resultData.data)).description()}
            case 'sale':
                return {doIndex:dind,desc:(new SaleData(resultData.data)).description()}
            case 'slant':
                return {doIndex:dind,desc:(new SlantData(resultData.data)).description()}
            case 'species':
                return {doIndex:dind,desc:(new SpeciesData(resultData.data)).description()}
            case 'sporePrint':
                return {doIndex:dind,desc:(new SporePrintData(resultData.data)).description()}
            case 'sporeSwab':
                return {doIndex:dind,desc:(new SporeSwabData(resultData.data)).description()}
            case 'stasisTube':
                return {doIndex:dind,desc:(new StasisTubeData(resultData.data)).description()}
            case 'subspecies':
                return {doIndex:dind,desc:(new SubspeciesData(resultData.data)).description()}
            case 'substrateBatch':
                return {doIndex:dind,desc:(new SubstrateBatchData(resultData.data)).description()}
            case 'substrateRecipe':
                return {doIndex:dind,desc:(new SubstrateRecipeData(resultData.data)).description()}
            case 'transfer':
                return {doIndex:dind,desc:(new TransferData(resultData.data)).description()}
            case 'user':
                return {doIndex:dind,desc:(new UserData(resultData.data)).description()}
            case 'waterJar':
                return {doIndex:dind,desc:(new WaterJarData(resultData.data)).description()}
            default:
                return {doIndex:dind,desc:"error: unknown item type"}
        }
    }).catch(e=>{
        return {doIndex:false,desc:"error: "+JSON.stringify(e)} // TODO: false ok here?
    })

    return {
        title: (itemType == "species" || itemType == "subspecies")?decoded:`${itemType} ${decoded}`,
        // title: { // TODO: ???
        //     absolute: decoded,
        // },
        description: await desc,
        alternates: {
            canonical: BaseExternalUrl+`/view/${itemType}/${idEncoded}`,
            languages: {
                "en-US": BaseExternalUrl+`/view/${itemType}/${idEncoded}`,
            }
        },
        robots: {
            index: doIndex,
            follow: true,
            // Optional finer control:
            nocache: true,
        },
    };
}

export interface ReactElementWithDoIndex {
    data: React.JSX.Element
    doIndex: boolean
}

const getDataResponse: (a1:string,a2:string,allCookies:string)=>Promise<Response> = async (itemTypeA: string, idEnc: string,allCookies:string) => {
    return new Promise<Response>((accept, reject) => {
        fetch(BaseExternalUrl + "/db/get/" + itemTypeA + "/" + idEnc, {
            method: 'Get',
            credentials: 'include',
            headers: {
                'Accept': 'application/json',
                //'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: ENSURE OK! maybe "*"?
                'Cookie': allCookies, // REQUIRED // TODO: can we drop this because we have included creds? TRY IT
                // TODO: set Origin header to web? or should this be BaseExternalUrl?
            },
        }).then(res=>{
            if (!res.ok) {
                res.text().then(txt=>{
                    reject("response not ok: " +  + ". Status " + res.status)
                }).catch(e=>{
                    reject(e)
                })
                return
            }
            accept(res)
        }).catch(err1 => {
            reject(err1)
        })
    })
}

const getData: (a1:string,a2:string,allCookies:string)=>Promise<{doIndex:boolean,data:any}> = async (itemTypeA: string, idEnc: string,allCookies:string) => {
    const res = await getDataResponse(itemTypeA,idEnc,allCookies)
    console.log("got response " + JSON.stringify(res))
    if (!res.ok) {
        throw new Error("response not ok: " + await res.text() + ". Status " + res.status)
    }
    return await getDataResponseToObject(res)
    // res.json().then((data) => {
    //     console.log(data)
    //     accept(data)
    // }).catch(err1 => {
    //     console.log("failed to resolve json data from result, " + JSON.stringify(err1))
    //     reject(err1)
    // })
}

async function getDataResponseToObject(res: Response):Promise<{doIndex:boolean,data:any}> {
    return await res.json().then((data) => {
        const doIndexHeader = res.headers.get('doIndex')
        console.log(data)
        return {doIndex: doIndexHeader==="true",data:data}
    }).catch(err1 => {
        console.log("failed to resolve json data from result, " + JSON.stringify(err1))
        throw err1
    })
}

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        itemType: string
        idEncoded: string // urlEncoded
    }>,
}) {
    const {itemType, idEncoded} = await params
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');

    try {
        const {doIndex, data} = await getData(itemType, idEncoded, allCookies)
        const readers = await GetReaderWriterNames() // Done on the server
        return <PageWrapper props={{pageType: "view", readers: readers}}>
            <CookiesProvider cookies={cookieStore.getAll()} session={session?.value}>
                    <MainViewArea itemType={itemType} inpData={data}/>
            </CookiesProvider>
        </PageWrapper>
    } catch (e) {
        return <PageWrapper props={{pageType: "view", readers: []}}>
                <div>{"Page not loaded. Nonexistent or unauthorized entry: "}</div>
                <div>{JSON.stringify(e)}</div>
        </PageWrapper>
    }

}

