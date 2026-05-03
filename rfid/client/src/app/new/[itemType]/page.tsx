import {GetReaderWriterNames} from "@/app/view/[itemType]/[idEncoded]/serverActions";
import {NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {NewAgarRecipeForm} from "@/app/components/agarRecipeClient";
import {NewBagForm} from "@/app/components/bagClient";
import {NewFruitingChamberForm} from "@/app/components/fruitingChamberClient";
import {NewJarForm} from "@/app/components/jarClient";
import {NewJarRecipeForm} from "@/app/components/jarRecipeClient";
import {NewLcForm} from "@/app/components/lcClient";
import {NewLcRecipeForm} from "@/app/components/lcRecipeClient";
import {NewPcRunForm} from "@/app/components/pcRunClient";
import {NewPlateForm} from "@/app/components/plateClient";
import {NewProjectForm} from "@/app/components/projectClient";
import {NewSaleForm} from "@/app/components/saleClient";
import {NewSlantForm} from "@/app/components/slantClient";
import {AssertSpecies, NewSpeciesForm} from "@/app/components/speciesClient";
import {NewStasisTubeForm} from "@/app/components/stasisTubeClient";
import {NewSubspeciesForm} from "@/app/components/subspeciesClient";
import {NewSubstrateRecipeForm} from "@/app/components/substrateRecipeClient";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import React from "react";
import PageWrapper from "@/app/components/clientGeneric";
import {NewSubstrateBatchForm} from "@/app/components/substrateBatchClient";
import {NewWaterJarForm} from "@/app/components/waterJarClient";
import {cookies} from "next/headers";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {BagData} from "@/app/components/bagServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {JarData} from "@/app/components/jarServer";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {LcData} from "@/app/components/lcServer";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {PcRunData} from "@/app/components/pcRunServer";
import {PlateData} from "@/app/components/plateServer";
import {ProjectData} from "@/app/components/projectServer";
import {SlantData} from "@/app/components/slantServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {WaterJarData} from "@/app/components/waterJarServer";
import {NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {ClientNewPage} from "@/app/new/[itemType]/client";

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
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');
    let speciesData: SpeciesData | undefined = undefined
    if (species !== undefined){
        speciesData = await fetch(BaseExternalUrl + "/db/get/" + species + "/" + species, {
            method: 'Get',
            credentials: 'include',
            headers: {
                'Accept': 'application/json',
                'Cookie': allCookies,
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
        <ClientNewPage itemType={itemType} species={speciesData}/>{/*fullPage class contained within*/}
    </PageWrapper>
}