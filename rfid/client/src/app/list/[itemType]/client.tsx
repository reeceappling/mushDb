'use client'

import {AssertPcRun, PcRunInline, PcRunListPageTable} from "@/app/components/pcRunClient";
import {PcRunData} from "@/app/components/pcRunServer";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import React from "react";
import {AssertArrayResult} from "@/app/components/common";
import {ListResult} from "@/app/components/formSubcomponents/shared";
import {
    AssertDualListResult,
    AssertSubstrateRecipe,
    SubstrateRecipeInline,
    SubstrateRecipeListPageTable
} from "@/app/components/substrateRecipeClient";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {AssertWaterJar, WaterJarInline, WaterJarListPageTable} from "@/app/components/waterJarClient";
import {WaterJarData} from "@/app/components/waterJarServer";
import {
    AssertSubstrateBatch,
    SubstrateBatchInline,
    SubstrateBatchListPageTable
} from "@/app/components/substrateBatchClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import SubspeciesDisplay, {
    AssertSubspecies,
    SubspeciesInline,
    SubspeciesListPageTable
} from "@/app/components/subspeciesClient";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {AssertStasisTube, StasisTubeInline, StasisTubeListPageTable} from "@/app/components/stasisTubeClient";
import {SporeSwab} from "@/app/components/sporeSwabServer";
import {AssertSporeSwab, SporeSwabInline, SporeSwabListPageTable} from "@/app/components/sporeSwabClient";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {AssertSporePrint, SporePrintInline, SporePrintListPageTable} from "@/app/components/sporePrintClient";
import {AssertSlant, SlantInline, SlantListPageTable} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {PlateData} from "@/app/components/plateServer";
import {AssertPlate, PlateInline, PlateListPageTable} from "@/app/components/plateClient";
import {MssData} from "@/app/components/mssServer";
import {AssertMss, MssInline, MssListPageTable} from "@/app/components/mssClient";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {AssertLcRecipe, LcRecipeInline, LcRecipeListPageTable} from "@/app/components/lcRecipeClient";
import {LcData} from "@/app/components/lcServer";
import {AssertLc, LcInline, LcListPageTable} from "@/app/components/lcClient";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {AssertJarRecipe, JarRecipeInline, JarRecipeListPageTable} from "@/app/components/jarRecipeClient";
import {LcSyringe} from "@/app/components/lcSyringeServer";
import {AssertLcSyringe, LcSyringeInline, LcSyringeListPageTable} from "@/app/components/lcSyringeClient";
import {JarData} from "@/app/components/jarServer";
import {AssertJar, JarInline, JarListPageTable} from "@/app/components/jarClient";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {
    AssertFruitingChamber,
    FruitingChamberInline,
    FruitingChamberListPageTable
} from "@/app/components/fruitingChamberClient";
import {FruitData} from "@/app/components/fruitServer";
import {AssertFruit, FruitInline, FruitListPageTable} from "@/app/components/fruitClient";
import {BagData} from "@/app/components/bagServer";
import {AssertBag, BagInline, BagListPageTable} from "@/app/components/bagClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {AgarBatchInline, AgarBatchListPageTable, AssertAgarBatch} from "@/app/components/agarBatchClient";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {AgarRecipeInline, AgarRecipeListPageTable, AssertAgarRecipe} from "@/app/components/agarRecipeClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {AssertGrainBatch, GrainBatchInline, GrainBatchListPageTable} from "@/app/components/grainBatchClient";
import {LatestListDisplay, LatestMostRecentListDisplay} from "@/app/components/clientGeneric";
import {AssertSpecies, SpeciesInline, SpeciesListPageTable} from "@/app/components/speciesClient";
import {SpeciesData} from "@/app/components/speciesServer";
import {ProjectData} from "@/app/components/projectServer";
import {AssertProject} from "@/app/components/projectClient";
import {SaleData} from "@/app/components/saleServer";
import {AssertSale, SaleInline, SaleListPageTable} from "@/app/components/saleClient";
import {TransferData} from "@/app/components/transferServer";
import {AssertTransfer} from "@/app/components/transferClient";
import {UserData} from "@/app/components/userServer";
import {AssertUser} from "@/app/components/userClient";

export default function ListDisplay({itemType,inpData}:{itemType: string, inpData: any}){
    try {
        switch (itemType) {
            case "agarBatches":
                AssertArrayResult<AgarBatchData>(inpData, AssertAgarBatch)
                return <AgarBatchListPageTable data={inpData} onClick={(v) => {
                        window.location.assign(BaseExternalUrl + "/view/agarBatch/" + encodeURI(v._id))}
                    } />
                // <div className={"listPageGrid"}>
                //     {inpData.map((v) => {
                //
                //     }
                //         return <AgarBatchListPageTable key={v._id} data={v} onClick={() => {
                //
                //         }}/>
                //     })}
                // </div>
                // // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                // //     <AgarBatchInline data={val} expandByDefault={false} onClick={() => {
                // //         window.location.assign(BaseExternalUrl + "/view/agarBatch/" + encodeURI(val._id))
                // //     }}/>
                // // }/>
            case "agarRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<AgarRecipeData>(inpData, AssertAgarRecipe)
                let arOc = (val: AgarRecipeData) => {
                    window.location.assign(BaseExternalUrl + "/view/agarRecipe/" + encodeURI(val._id))}
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <AgarRecipeListPageTable data={inpData.recent || []} onClick={arOc}/> {/* TODO: recipe id???*/}
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <AgarRecipeListPageTable data={inpData.standard || []} onClick={arOc}/> {/* TODO: recipe id???*/}
                </>

                // return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                //     return <AgarRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/agarRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                // }}/>
            case "bags":
                AssertArrayResult<BagData>(inpData, AssertBag)
                return <BagListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/bag/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <BagInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/bag/" + encodeURI(val._id)) // TODO: url ok?
                //     }}/>
                // }/>
            case "fruits":
                AssertArrayResult<FruitData>(inpData, AssertFruit)
                return <FruitListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/fruit/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <FruitInline data={val} idIsLink={true} showMainPageButton={true} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/fruit/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "fruitingChambers":
                AssertArrayResult<FruitingChamberData>(inpData, AssertFruitingChamber)
                return <FruitingChamberListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/fruitingChamber/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     // TODO: DISPLAYING TOO MUCH
                //     <FruitingChamberInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/fruitingChamber/" + encodeURI(val._id)) // TODO: url ok?
                //     }}/>
                // }/>
            case "jars":
                AssertArrayResult<JarData>(inpData, AssertJar)
                return <JarListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/jar/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <JarInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/jar/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "jarRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<JarRecipeData>(inpData, AssertJarRecipe)
                let jrOc = (val: JarRecipeData) => {
                    window.location.assign(BaseExternalUrl + "/view/jarRecipe/" + encodeURI(val._id))}
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <JarRecipeListPageTable data={inpData.recent || []} onClick={jrOc}/>
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <JarRecipeListPageTable data={inpData.standard || []} onClick={jrOc}/>
                </>
                // return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                //     return <JarRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/jarRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                // }}/>
            case "grainBatches": // TODO: validate works as expected
                AssertArrayResult<GrainBatchData>(inpData, AssertGrainBatch)
                return <GrainBatchListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/grainBatch/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=> {
                //     return <GrainBatchInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/grainBatch/" + encodeURI(val._id))}}/>
                // }}/>
            case "lcs":
                AssertArrayResult<LcData>(inpData, AssertLc)
                return <LcListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/lc/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <LcInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/lc/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "lcSyringes":
                AssertArrayResult<LcSyringe>(inpData, AssertLcSyringe)
                return <LcSyringeListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/lcSyringe/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <LcSyringeInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/lcSyringe/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "lcRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<LcRecipeData>(inpData, AssertLcRecipe)
                let lcrOc = (val: LcRecipeData) => {
                    window.location.assign(BaseExternalUrl + "/view/lcRecipe/" + encodeURI(val._id))}
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <LcRecipeListPageTable data={inpData.recent || []} onClick={lcrOc}/> {/* TODO: recipe id???*/}
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <LcRecipeListPageTable data={inpData.standard || []} onClick={lcrOc}/> {/* TODO: recipe id???*/}
                </>
                // return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                //     return <LcRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/lcRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                // }}/>
            case "mss":
                AssertArrayResult<MssData>(inpData, AssertMss)
                return <MssListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/mss/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <MssInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/mss/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "pcRuns":
                AssertArrayResult<PcRunData>(inpData, AssertPcRun)
                return <PcRunListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/pcRun/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <PcRunInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/pcRun/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "plates":
                AssertArrayResult<PlateData>(inpData, AssertPlate)
                return <PlateListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/plate/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <PlateInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/plate/" + encodeURI(val._id))
                //     }}/>
                // }/>
            // // case "plugs": // TODO: THIS
            // //     return <PlugsDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            // case "projects":// TODO: TEST
            //     AssertArrayResult<ProjectData>(inpData, AssertProject)
            //     return <LatestListDisplay data={inpData} constructor={(val, i)=>
            //         <ProjectInline data={val} expandByDefault={false} onClick={()=>{window.location.assign(BaseExternalUrl+"/view/project/"+encodeURI(val._id))}}/>
            //     }/>
            case "sales":// TODO: TEST
                AssertArrayResult<SaleData>(inpData, AssertSale)
                return <SaleListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/sale/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <SaleInline data={val} expandByDefault={false} onClick={()=>{window.location.assign(BaseExternalUrl+"/view/sale/"+encodeURI(val._id))}}/>
                // }/>
            case "slants":
                AssertArrayResult<SlantData>(inpData, AssertSlant)
                return <SlantListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/slant/" + encodeURI(v._id))}
                } />
            case "species":// TODO: TEST
                AssertArrayResult<SpeciesData>(inpData, AssertSpecies)
                return <SpeciesListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/species/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <SpeciesInline data={val} expandByDefault={false} onClick={()=>{
                //         console.log("redirecting to: "+BaseExternalUrl+"/view/species/"+encodeURI(val._id))
                //         window.location.assign(BaseExternalUrl+"/view/species/"+encodeURI(val._id))}
                //     }/>
                // }/>
            case "sporePrints":
                AssertArrayResult<SporePrintData>(inpData, AssertSporePrint)
                return <SporePrintListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/sporePrint/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <SporePrintInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/sporePrint/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "sporeSwabs": // TODO: TEST after SporeSwabInline
                AssertArrayResult<SporeSwab>(inpData, AssertSporeSwab)
                return <SporeSwabListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/sporeSwab/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <SporeSwabInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/sporeSwab/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "stasisTubes":
                AssertArrayResult<StasisTubeData>(inpData, AssertStasisTube)
                return <StasisTubeListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/stasisTube/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <StasisTubeInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/stasisTube/" + encodeURI(val._id))
                //     }}/>
                // }/>
            case "subspecies": // TODO: test functionality. Unsure if we even want to be able to list these
                AssertArrayResult<SubspeciesData>(inpData, AssertSubspecies)
                return <SubspeciesListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/subspecies/" + encodeURI(v._id))} // TODO: ensure works
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=> {
                //     return <SubspeciesInline key={i} showSpeciesName={true} props={{data:val, expandByDefault:false, onClick:() => { // TODO: showSpeciesName true ok?
                //         window.location.assign(BaseExternalUrl + "/view/subspecies/" + encodeURI(val._id))}}} /> // TODO: url ok?
                // }}/>
            case "substrateRecipes":
                // TODO: allow lookup or navigation by recipe name
                AssertDualListResult<SubstrateRecipeData>(inpData, AssertSubstrateRecipe)
                let subrOc = (val: SubstrateRecipeData) => {
                    window.location.assign(BaseExternalUrl + "/view/substrateRecipe/" + encodeURI(val._id))}
                return <>
                    <div className={"text-xl centerH"}>{"Recent"}</div>
                    <SubstrateRecipeListPageTable data={inpData.recent || []} onClick={subrOc}/>
                    <div className={"text-xl centerH"}>{"Standard"}</div>
                    <SubstrateRecipeListPageTable data={inpData.standard || []} onClick={subrOc}/>
                </>
                // return <LatestMostRecentListDisplay data={inpData} constructor={(val, i)=> {
                //     return <SubstrateRecipeInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/substrateRecipe/" + encodeURI(val._id))}}/> // TODO: recipe id???
                // }}/>
            case "substrateBatches":
                AssertArrayResult<SubstrateBatchData>(inpData, AssertSubstrateBatch)
                return <SubstrateBatchListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/substrateBatch/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <SubstrateBatchInline data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/substrateBatch/" + encodeURI(val._id)) // TODO: ensure url ok
                //     }}/>
                // }/>
            // case "transfers":
            //     AssertArrayResult<TransferData>(inpData, AssertTransfer)
            // TODO: new page table format
            //     return <LatestListDisplay data={inpData} constructor={(val, i)=>
            //         <TransferInline data={val} expandByDefault={false} onClick={() => {
            //             window.location.assign(BaseExternalUrl + "/view/transfer/" + encodeURI(val._id)) // TODO: ensure url ok
            //         }}/>
            //     }/>
            // case "users":
            //     AssertArrayResult<UserData>(inpData, AssertUser)
            // TODO: new page table format
            //     return <LatestListDisplay data={inpData} constructor={(val, i)=>
            //         <UserInline data={val} expandByDefault={false} onClick={() => {
            //             window.location.assign(BaseExternalUrl + "/view/user/" + encodeURI(val._id)) // TODO: ensure url ok
            //         }}/>
            //     }/>
            // case "users": // TODO: this
            //     return <UserDisplay data={inpData} readonly={false} id={id} isTopLevel={true} headerLevel={TopPageHeaderLevel} cookies={allCookies}/>// TODO: FIX!
            case "waterJars":
                AssertArrayResult<WaterJarData>(inpData, AssertWaterJar)
                return <WaterJarListPageTable data={inpData} onClick={(v) => {
                    window.location.assign(BaseExternalUrl + "/view/waterJar/" + encodeURI(v._id))}
                } />
                // return <LatestListDisplay data={inpData} constructor={(val, i)=>
                //     <WaterJarInline key={i} data={val} expandByDefault={false} onClick={() => {
                //         window.location.assign(BaseExternalUrl + "/view/waterJar/" + encodeURI(val._id))
                //     }}/>
                // }/>
            default:
                return <ErrorDisplay err={"Unhandled list item type: " + itemType} headerLevel={TopPageHeaderLevel}/>
        }
    } catch (error) {
        return <ErrorDisplay err={JSON.stringify(error)} headerLevel={1}></ErrorDisplay>
    }
}