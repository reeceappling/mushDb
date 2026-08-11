import React from "react";
import {BaseExternalUrl, mushDbTitle} from "@/app/components/Constants";
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

type Props = {
    params: Promise<{
        itemType: string
        idEncoded: string // urlEncoded
    }>
};
// Next.js runs this first to set the tab title
export async function generateMetadata({ params }: Props): Promise<Metadata> { // TODO: add generateMetadata on all pages!
    const {itemType, idEncoded} = await params
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
    const decoded = decodeURI(idEncoded)
    let desc = getData(itemType, idEncoded, allCookies).then(resultData=>{
        switch(itemType){
            case 'agarBatch':
                return (new AgarBatchData(resultData)).description()
            case 'agarRecipe':
                return (new AgarRecipeData(resultData)).description()
            case 'bag':
                return (new BagData(resultData)).description()
            case 'fruit':
                return (new FruitData(resultData)).description()
            case 'fruitingChamber':
                return (new FruitingChamberData(resultData)).description()
            case 'grainBatch':
                return (new GrainBatchData(resultData)).description()
            case 'jar':
                return (new JarData(resultData)).description()
            case 'jarRecipe':
                return (new JarRecipeData(resultData)).description()
            case 'lc':
                return (new LcData(resultData)).description()
            case 'lcRecipe':
                return (new LcRecipeData(resultData)).description()
            case 'lcSyringe':
                return (new LcSyringeData(resultData)).description()
            case 'mss':
                return (new MssData(resultData)).description()
            case 'pcRun':
                return (new PcRunData(resultData)).description()
            case 'plate':
                return (new PlateData(resultData)).description()
            case 'plugs':
                return (new PlugsData(resultData)).description()
            case 'project':
                return (new ProjectData(resultData)).description()
            case 'sale':
                return (new SaleData(resultData)).description()
            case 'slant':
                return (new SlantData(resultData)).description()
            case 'species':
                return (new SpeciesData(resultData)).description()
            case 'sporePrint':
                return (new SporePrintData(resultData)).description()
            case 'sporeSwab':
                return (new SporeSwabData(resultData)).description()
            case 'stasisTube':
                return (new StasisTubeData(resultData)).description()
            case 'subspecies':
                return (new SubspeciesData(resultData)).description()
            case 'substrateBatch':
                return (new SubstrateBatchData(resultData)).description()
            case 'substrateRecipe':
                return (new SubstrateRecipeData(resultData)).description()
            case 'transfer':
                return (new TransferData(resultData)).description()
            case 'user':
                return (new UserData(resultData)).description()
            case 'waterJar':
                return (new WaterJarData(resultData)).description()
            default:
                return "error: unknown item type"
        }
    }).catch(e=>{
        return "error: "+JSON.stringify(e)
    })

    if (itemType == "species" || itemType == "subspecies") {
        return {
            title: decoded,
            // title: { // TODO: ???
            //     absolute: decoded,
            // },
            description: await desc,
        };
    }
    return {
        title: itemType+` `+decoded,
        description: await desc,
    };
}

const getData: (a1:string,a2:string,allCookies:string)=>Promise<any> = async (itemTypeA: string, idEnc: string,allCookies:string) => {
    return new Promise<React.JSX.Element>((accept, reject) => { // TODO: REIMPLEMENT!
        fetch(BaseExternalUrl + "/db/get/" + itemTypeA + "/" + idEnc, {
            method: 'Get',
            credentials: 'include',
            headers: {
                'Accept': 'application/json',
                //'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: ENSURE OK! maybe "*"?
                'Cookie': allCookies, // REQUIRED // TODO: can we drop this because we have included creds?
                // TODO: set Origin header to web? or should this be BaseExternalUrl?
            },
        }).then((res) => {
            console.log("got response " + JSON.stringify(res))
            if (!res.ok) {
                return res.text().then(txt => {
                    throw new Error("response not ok: " + txt + ". Status " + res.status)
                }).catch(err => {
                    throw new Error("response not ok and failed to decode: " + JSON.stringify(err) + ". Status " + res.status)
                })
            }
            res.json().then((data) => {
                console.log(data)
                accept(data)
            }).catch(err1 => {
                console.log("failed to resolve json data from result, " + JSON.stringify(err1))
                reject(err1)
            })
        }).catch(err1 => {
            reject(err1)
        })
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
        const data = await getData(itemType, idEncoded, allCookies)
        const readers = await GetReaderWriterNames()
        return <PageWrapper props={{pageType: "view", readers: readers}}>
            <CookiesProvider cookies={cookieStore.getAll()} session={session?.value}> {/* TODO: validate working*/}
                {/*<div className={"fullPage"}>*/}
                    <MainViewArea itemType={itemType} inpData={data}/>
                {/*</div>*/}
            </CookiesProvider>
        </PageWrapper>
    } catch (e) {
        return <PageWrapper props={{pageType: "view", readers: []}}>
            {/*<div className={"fullPage"}>*/}
                <div>{"Page not loaded. Nonexistent or unauthorized entry: "}</div> {/* TODO: STYLING*/}
                <div>{JSON.stringify(e)/* TODO: CHANGE!*/}</div>
            {/*</div>*/}
        </PageWrapper>
    }

}

