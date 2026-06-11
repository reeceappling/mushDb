'use client'
import {SpeciesData} from "@/app/components/speciesServer";
import {NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {NewAgarRecipeForm} from "@/app/components/agarRecipeClient";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {NewBagForm} from "@/app/components/bagClient";
import {BagData} from "@/app/components/bagServer";
import {NewFruitingChamberForm} from "@/app/components/fruitingChamberClient";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {NewJarForm} from "@/app/components/jarClient";
import {JarData} from "@/app/components/jarServer";
import {NewJarRecipeForm} from "@/app/components/jarRecipeClient";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {NewLcForm} from "@/app/components/lcClient";
import {LcData} from "@/app/components/lcServer";
import {NewLcRecipeForm} from "@/app/components/lcRecipeClient";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {NewPcRunForm} from "@/app/components/pcRunClient";
import {PcRunData} from "@/app/components/pcRunServer";
import {NewPlateForm} from "@/app/components/plateClient";
import {PlateData} from "@/app/components/plateServer";
import {NewProjectForm} from "@/app/components/projectClient";
import {ProjectData} from "@/app/components/projectServer";
import {NewSlantForm} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {NewSpeciesForm} from "@/app/components/speciesClient";
import {NewStasisTubeForm} from "@/app/components/stasisTubeClient";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {NewSubspeciesForm} from "@/app/components/subspeciesClient";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {NewSubstrateRecipeForm} from "@/app/components/substrateRecipeClient";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {NewSubstrateBatchForm} from "@/app/components/substrateBatchClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {NewWaterJarForm} from "@/app/components/waterJarClient";
import {WaterJarData} from "@/app/components/waterJarServer";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {TopPageHeaderLevel} from "@/app/components/Constants";
import React, {JSX, useState} from "react";
import EntryLinkForId from "@/app/components/formSubcomponents/entryLink";
import { NewPlugsForm } from "@/app/components/plugsClient";
import {PlugsData} from "@/app/components/plugsServer";

export function ClientNewPage({itemType, species}: { itemType: string, species?: SpeciesData }) {
    const [newEntries, setNewEntries] = useState<JSX.Element[]>([])
    const createdLinkFor = (linkText: string, linkId: string, typ: string) => {
        return <EntryLinkForId props={{openInNewTab:false/* TODO: ok?*/,displayId: linkText, linkId: linkId, entryType: typ}}/>
    }
    const createdItemsArea = ()=>{
        return <div>
            {/* TODO: styling! */}
            {newEntries}
        </div>
    }
    const newForm = () => {
        switch (itemType) {
            // case "agarBatch": // only from AgarRecipes page?
            //     return <NewAgarBatchForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: AgarBatchData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "agarBatch")])
            //         }
            //     }}/>
            case "agarRecipe": // from self only
                return <NewAgarRecipeForm handlers={{
                    // TODO: redirect?
                    isTopLevel: true, onCreate: (newEntry: AgarRecipeData) => {
                        setNewEntries([...newEntries,createdLinkFor(newEntry.name, newEntry._id, "agarRecipe")])
                    }
                }}/>
            // case "bag": // only from the substrateBatch page?
            //     return <NewBagForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: BagData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "bag")])
            //         }
            //     }}/>
            // // Fruit is not available for new. Only made in bag/box pages
            // case "fruitingChamber": // from either lc, jar, or substrateBatch
            //     return <NewFruitingChamberForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: FruitingChamberData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "fruitingChamber")])
            //         }
            //     }}/>
            // case "grainBatch": // from either lc, jar, or substrateBatch
            //     return <NewGrainBatchForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: GrainBatchData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "grainBatch")])
            //         }
            //     }}/>
            // case "jar": // from jar recipe only (or pcRun?)
            //     return <NewJarForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: JarData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "jar")])
            //         }
            //     }}/>
            case "jarRecipe": // from self only
                return <NewJarRecipeForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: JarRecipeData) => {
                        // TODO: name ok here?
                        setNewEntries([...newEntries,createdLinkFor(newEntry.name, newEntry._id, "jarRecipe")])
                    }
                }}/>
            // case "lc": // from lcRecipe
            //     return <NewLcForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: LcData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "lc")])
            //         }
            //     }}/>
            case "lcRecipe": // from self only
                return <NewLcRecipeForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: LcRecipeData) => {
                        // TODO: name ok here?
                        setNewEntries([...newEntries,createdLinkFor(newEntry.name, newEntry._id, "lcRecipe")])
                    }
                }}/>
            // MSS only built-in to sporePrint page?
            case "pcRun": // from self only, or embedded in other creators
                return <NewPcRunForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: PcRunData) => {
                        setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "pcRun")])
                    }
                }}/>
            // case "plate": // from agarBatch only
            //     // TODO: above newPlateForm (and all others), put links area for created items
            //     return <NewPlateForm handlers={{
            //         isTopLevel: true,
            //         onCreate: (newEntry: PlateData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "plate")])
            //         }
            //     }}/>
            case "plugs":
                return <NewPlugsForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: PlugsData) => {
                        setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "plugs")])
                    }
                }}/>
            case "project": // from this page only
                return <NewProjectForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: ProjectData) => {
                        setNewEntries([...newEntries,createdLinkFor(newEntry._id, encodeURI(newEntry._id), "project")])
                    },
                }}/> //TODO: REENABLE??? onCreate={pd => doRedirect(pd._id)}
            // case "sale": // TODO: from other pages only!
            //     return <NewSaleForm headerLevel={TopPageHeaderLevel} onCreate={sd => doRedirect(sd._id)}/>
            // case "slant": // from agarBatch only
            //     return <NewSlantForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: SlantData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "slant")])
            //         }
            //     }}/>
            case "species": // from this page only
                return <NewSpeciesForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: SpeciesData) => {
                        // TODO: urlencode? use normal or scientific name?
                        setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "species")])
                    }
                }}/>
            // case "sporePrint": // TODO: only from fruit page or new page
            //     return <NewSporePrintForm redirectOnCreate={true}/>
            // case "sporeSwab": // TODO: SporeSwab only from sporePrint page or new page
            //     return <NewSporeSwabForm redirectOnCreate={true}/>
            // case "stasisTube": // TODO: from where? pcRun?
            //     return <NewStasisTubeForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: StasisTubeData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "stasisTube")])
            //         }
            //     }}/>
            case "subspecies": // TODO: from species page or own page
                // TODO: ERROR IF SPECIES NOT FOUND
                return <NewSubspeciesForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: SubspeciesData) => {
                        // TODO: id here or no?
                        setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "subspecies")])
                    }
                }} species={species}/> // TODO: ensure speciesData works here
            case "substrateRecipe": // from this page only, or embedded pages
                return <NewSubstrateRecipeForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: SubstrateRecipeData) => {
                        // TODO: name here ok?
                        setNewEntries([...newEntries,createdLinkFor(newEntry.name, newEntry._id, "substrateRecipe")])
                    }
                }}/>
            // case "substrateBatch": // TODO: from recipe page only, or embedded?
            //     return <NewSubstrateBatchForm handlers={{
            //         isTopLevel: true, onCreate: (newEntry: SubstrateBatchData) => {
            //             setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "substrateBatch")])
            //         }
            //     }}/>
            case "waterJar":
                return <NewWaterJarForm handlers={{
                    isTopLevel: true, onCreate: (newEntry: WaterJarData) => {
                        setNewEntries([...newEntries,createdLinkFor(newEntry._id, newEntry._id, "waterJar")])
                    }
                }}/>
            // Transfers only made in other pages
            default:
                return <ErrorDisplay err={"Invalid new item type in path: " + itemType}
                                     headerLevel={TopPageHeaderLevel}/>
        }
    }
    return <div className={"fullPage"}>
        {createdItemsArea()}
        {newForm()}
    </div>
}