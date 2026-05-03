'use client'

import {AssertPcRun, PcRunInline} from "@/app/components/pcRunClient";
import {PcRunData} from "@/app/components/pcRunServer";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import React from "react";
import {AssertArrayResult} from "@/app/components/common";
import {ListResult} from "@/app/components/formSubcomponents/shared";
import {AssertDualListResult, AssertSubstrateRecipe, SubstrateRecipeInline} from "@/app/components/substrateRecipeClient";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {AssertWaterJar, WaterJarInline} from "@/app/components/waterJarClient";
import {WaterJarData} from "@/app/components/waterJarServer";
import {AssertSubstrateBatch, SubstrateBatchInline} from "@/app/components/substrateBatchClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import SubspeciesDisplay, {AssertSubspecies, SubspeciesInline} from "@/app/components/subspeciesClient";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {AssertStasisTube, StasisTubeInline} from "@/app/components/stasisTubeClient";
import {SporeSwab} from "@/app/components/sporeSwabServer";
import {AssertSporeSwab, SporeSwabInline} from "@/app/components/sporeSwabClient";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {AssertSporePrint, SporePrintInline} from "@/app/components/sporePrintClient";
import {AssertSlant, SlantInline} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {PlateData} from "@/app/components/plateServer";
import {AssertPlate, PlateInline} from "@/app/components/plateClient";
import {MssData} from "@/app/components/mssServer";
import {AssertMss, MssInline} from "@/app/components/mssClient";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {AssertLcRecipe, LcRecipeInline} from "@/app/components/lcRecipeClient";
import {LcData} from "@/app/components/lcServer";
import {AssertLc, LcInline} from "@/app/components/lcClient";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {AssertJarRecipe, JarRecipeInline} from "@/app/components/jarRecipeClient";
import {LcSyringe} from "@/app/components/lcSyringeServer";
import {AssertLcSyringe, LcSyringeInline} from "@/app/components/lcSyringeClient";
import {JarData} from "@/app/components/jarServer";
import {AssertJar, JarInline} from "@/app/components/jarClient";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {AssertFruitingChamber, FruitingChamberInline} from "@/app/components/fruitingChamberClient";
import {FruitData} from "@/app/components/fruitServer";
import {AssertFruit, FruitInline} from "@/app/components/fruitClient";
import {BagData} from "@/app/components/bagServer";
import {AssertBag, BagInline} from "@/app/components/bagClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {AgarBatchInline, AssertAgarBatch} from "@/app/components/agarBatchClient";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {AgarRecipeInline, AssertAgarRecipe} from "@/app/components/agarRecipeClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {AssertGrainBatch, GrainBatchInline} from "@/app/components/grainBatchClient";
import {LatestListDisplay, LatestMostRecentListDisplay} from "@/app/components/clientGeneric";

export default function ListDisplay({itemType,inpData}:{itemType: string, inpData: any}){
    try {
        switch (itemType) {
            case "agarBatches":
                AssertArrayResult<AgarBatchData>(inpData, AssertAgarBatch)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <AgarBatchInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/agarBatch/" + encodeURI(val._id))
                    }}/>
                }/>
            case "agarRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<AgarRecipeData>(inpData, AssertAgarRecipe)
                return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                    return <AgarRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/agarRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                }}/>
            case "bags":
                AssertArrayResult<BagData>(inpData, AssertBag)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <BagInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/bag/" + encodeURI(val._id)) // TODO: url ok?
                    }}/>
                }/>
            case "fruits":
                AssertArrayResult<FruitData>(inpData, AssertFruit)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <FruitInline data={val} idIsLink={true} showMainPageButton={true} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/fruit/" + encodeURI(val._id))
                    }}/>
                }/>
            case "fruitingChambers":
                AssertArrayResult<FruitingChamberData>(inpData, AssertFruitingChamber)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    // TODO: DISPLAYING TOO MUCH
                    <FruitingChamberInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/fruitingChamber/" + encodeURI(val._id)) // TODO: url ok?
                    }}/>
                }/>
            case "jars":
                AssertArrayResult<JarData>(inpData, AssertJar)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <JarInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/jar/" + encodeURI(val._id))
                    }}/>
                }/>
            case "jarRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<JarRecipeData>(inpData, AssertJarRecipe)
                return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                    return <JarRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/jarRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                }}/>
            case "grainBatches": // TODO: validate works as expected
                AssertArrayResult<GrainBatchData>(inpData, AssertGrainBatch)
                return <LatestListDisplay data={inpData} constructor={(val, i)=> {
                    return <GrainBatchInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/grainBatch/" + encodeURI(val._id))}}/>
                }}/>
            case "lcs":
                AssertArrayResult<LcData>(inpData, AssertLc)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <LcInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/lc/" + encodeURI(val._id))
                    }}/>
                }/>
            case "lcSyringes":
                AssertArrayResult<LcSyringe>(inpData, AssertLcSyringe)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <LcSyringeInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/lcSyringe/" + encodeURI(val._id))
                    }}/>
                }/>
            case "lcRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<LcRecipeData>(inpData, AssertLcRecipe)
                return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                    return <LcRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/lcRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                }}/>
            case "mss":
                AssertArrayResult<MssData>(inpData, AssertMss)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <MssInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/mss/" + encodeURI(val._id))
                    }}/>
                }/>
            case "pcRuns":
                AssertArrayResult<PcRunData>(inpData, AssertPcRun)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <PcRunInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/pcRun/" + encodeURI(val._id))
                    }}/>
                }/>
            case "plates":
                AssertArrayResult<PlateData>(inpData, AssertPlate)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <PlateInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/plate/" + encodeURI(val._id))
                    }}/>
                }/>
            // // case "plugs": // TODO: THIS
            // //     return <PlugsDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            // case "projects":
            //     return <ProjectDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            // case "sales":
            //     return <SaleDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            case "slants":
                AssertArrayResult<SlantData>(inpData, AssertSlant)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <SlantInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/slant/" + encodeURI(val._id))
                    }}/>
                }/>
            // case "species":// TODO: TEST
            //     //return <SpeciesInline data={} expandByDefault={false} onClick={()=>{window.location.assign(BaseExternalUrl+"/view/species/"+entry._id)}}
            //     return <SpeciesDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            case "sporePrints":
                AssertArrayResult<SporePrintData>(inpData, AssertSporePrint)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <SporePrintInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/sporePrint/" + encodeURI(val._id))
                    }}/>
                }/>
            case "sporeSwabs": // TODO: TEST after SporeSwabInline
                AssertArrayResult<SporeSwab>(inpData, AssertSporeSwab)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <SporeSwabInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/sporeSwab/" + encodeURI(val._id))
                    }}/>
                }/>
            case "stasisTubes":
                AssertArrayResult<StasisTubeData>(inpData, AssertStasisTube)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <StasisTubeInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/stasisTube/" + encodeURI(val._id))
                    }}/>
                }/>
            case "subspecies": // TODO: test functionality. Unsure if we even want to be able to list these
                // TODO: NOT WORKING, LIKELY NOT A DUAL LIST
                return <ErrorDisplay err={"NOT IMPLEMENTED"} />
                // AssertDualListResult<SubspeciesData>(inpData, AssertSubspecies)
                // return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                //     return <SubspeciesInline key={i} showSpeciesName={true} props={{data:val, expandByDefault:false, onClick:() => { // TODO: showSpeciesName true ok?
                //         window.location.assign(BaseExternalUrl + "/view/subspecies/" + encodeURI(val._id))}}} /> // TODO: url ok?
                // }}/>
            case "substrateRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<SubstrateRecipeData>(inpData, AssertSubstrateRecipe)
                return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                    return <SubstrateRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/substrateRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
            }}/>
            case "substrateBatches":
                AssertArrayResult<SubstrateBatchData>(inpData, AssertSubstrateBatch)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <SubstrateBatchInline data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/substrateBatch/" + encodeURI(val._id)) // TODO: ensure url ok
                    }}/>
                }/>
            // case "transfers":
            //     return <TransferDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            // case "users": // TODO: this
            //     return <UserDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            case "waterJars":
                AssertArrayResult<WaterJarData>(inpData, AssertWaterJar)
                return <LatestListDisplay data={inpData} constructor={(val, i)=>
                    <WaterJarInline key={i} data={val} expandByDefault={false} onClick={() => {
                        window.location.assign(BaseExternalUrl + "/view/waterJar/" + encodeURI(val._id))
                    }}/>
                }/>
            default:
                return <ErrorDisplay err={"Unhandled list item type: " + itemType} headerLevel={TopPageHeaderLevel}/>
        }
    } catch (error) {
        return <ErrorDisplay err={JSON.stringify(error)} headerLevel={1}></ErrorDisplay>
    }
}