'use client'

import {JSX, useContext, useEffect, useState} from "react";
import {Liquid} from "./liquids";
import EntryLinkForId, {
    EntryLinkIdWrapper
} from "@/app/components/formSubcomponents/entryLink";
import {AllEntries, Data, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {
    ImageLocationFor,
    NewPicWithNotesForm,
    PicWithNotesForm,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {PixRows} from "@/app/components/formSubcomponents/commonClient2";
import {
    InputDecimal,
    InputText,
    InputTextInlineTitle,
    InputTextWithSmallTitle,
    NumericalAreaWithAbsolutes
} from "./numericInput";
import DateArea, {NumberToDate} from "./date";
import {Note, NotesAreaMostRecentImage, NotesFormArea, SingleNoteV2} from "./notes";
import {SpeciesData} from "@/app/components/speciesServer";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {NoSsr} from "@mui/material";
import {useQuery} from "@tanstack/react-query";
import {dataFor, viewUrlFor} from "@/app/components/common";
import {SelectorFor} from "@/app/components/selector";
import TextBoxArea from "@/app/components/formSubcomponents/singleTextBoxArea";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {DepthContext, DepthProvider} from "./depthContext/depth";
import {DowelType} from "@/app/components/plugsServer";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import TestAndValidate from "@/app/components/testing/untested";
import * as React from "react";

// export function OnClickWrapper(props: React.PropsWithChildren<{ handleClick?: () => void }>) {
//     return <div className={"hoverClickable"} onClick={(e) => {
//         e.stopPropagation() // TODO: ok?
//         e.preventDefault() // TODO: ok?
//         props.handleClick && props.handleClick()
//     }}>{props.children}</div>
// }

export function RemoveToggle({disabled, click, keptClass, removedClass, keptTxt, removedTxt}: {
    disabled: boolean,
    click: () => void,
    keptClass: string,
    removedClass?: string,
    keptTxt: string,
    removedTxt?: string
}) {
    return <button className={disabled ? removedClass : keptClass}
                   onClick={(e) => {
                       e.stopPropagation();
                       click();
                   }}>{disabled ? removedTxt : keptTxt}</button>
}

export function RemoveButton({txt, click}: { click: () => void, txt: string }) {
    return <button className={"removeButtonSmall"} onClick={(e) => {
        e.stopPropagation();
        click();
    }}>{txt}</button>
}


export function LiquidEntryForNew({initial, updateParent}: {
    initial: Liquid,
    updateParent: (l: Liquid) => void,
}) {
    const [current, setCurrent] = useState<Liquid>(initial);
    const handleFormChangePct = (pct: number) => {
        updateParent({...structuredClone(current), pct: pct})
    }
    useEffect(() => {
        setCurrent(initial)
    }, [initial]);
    return <>
        <div className={"text-m"}>{current.name}</div>
        <InputDecimal label={"Percentage by volume"} initial={initial.pct} updateParent={handleFormChangePct}  min={0.0} max={100}/>{/* TODO: min/max ok?*/}
    </>
}

export function NutrientEntryForNew({initial, updateParent}: {
    initial: Nutrient,
    updateParent: (l: Nutrient) => void
}) {
    const [current, setCurrent] = useState<Nutrient>(initial)
    const [errTxt, setErrTxt] = useState<string | undefined>()
    useEffect(() => {
        setCurrent(initial)
    }, [initial]);
    const handleFormChangeAmt = (val: number) => {
        const data = {...current};
        data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        const data = {...current};
        data.unit = val
        updateParent(data)
    }
    return <>
        <div className={"text-m"}>{current.nutrient}</div>
        <InputDecimal label={"Amount"} initial={initial.amount} updateParent={handleFormChangeAmt} min={0.0} max={1000.0}/>{/* TODO: min/max ok?*/}
        {/*<NumericalAreaWithAbsolutes label="Amount" mode="floating" min={0.0} max={100.0} readonly={false}*/}
        {/*                            errorMessage={err} value={currentValue.amount.toString()}*/}
        {/*                            onChange={(val?: string) => {*/}
        {/*                                try {*/}
        {/*                                    const n = Number(val) // TODO: allow only numbers here*/}
        {/*                                    if (Number.isNaN(n)) {*/}
        {/*                                        setErr("NaN input")*/}
        {/*                                    } else {*/}
        {/*                                        val && handleFormChangeAmt(n)*/}
        {/*                                        setErr(undefined)*/}
        {/*                                    }*/}
        {/*                                } catch (e) {*/}
        {/*                                    setErr(JSON.stringify(e))*/}
        {/*                                }*/}
        {/*                            }}/>*/}
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt}
                                 value={current.unit.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                console.error(JSON.stringify(e))
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function SugarEntryForNew({initial, updateParent}: {
    initial: Sugar,
    updateParent: (l: Sugar) => void
}) {
    const [errTxt, setErrTxt] = useState<string | undefined>()
    const [current, setCurrent] = useState<Sugar>(initial)
    useEffect(() => {
        setCurrent(initial)
    }, [initial]);
    const handleFormChangeAmt = (val: number) => {
        const data = {...current};
        data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        const data = {...current};
        data.unit = val
        updateParent(data)
    }
    return <>
        <div className={"text-m"}>{current.type}</div>
        <InputDecimal label={"Amount"} initial={initial.amount} updateParent={handleFormChangeAmt}  min={0.0} max={1000.0}/>{/* TODO: min/max ok?*/}
        {/*<NumericalAreaWithAbsolutes label="Amount" mode="floating" min={0.0} max={100.0} readonly={false}*/}
        {/*                            errorMessage={err} value={current.amount.toString()}*/}
        {/*                            onChange={(val?: string) => {*/}
        {/*                                try {*/}
        {/*                                    const n = Number(val) // TODO: allow only numbers here*/}
        {/*                                    if (Number.isNaN(n)) {*/}
        {/*                                        setErr("NaN input")*/}
        {/*                                    } else {*/}
        {/*                                        val && handleFormChangeAmt(n)*/}
        {/*                                        setErr(undefined)*/}
        {/*                                    }*/}
        {/*                                } catch (e) {*/}
        {/*                                    setErr(JSON.stringify(e))*/}
        {/*                                }*/}
        {/*                            }}/>*/}
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt}
                                 value={current.unit.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function AdditiveEntryForNew({initial, updateParent}: {
    initial: Additive,
    updateParent: (l: Additive) => void
}) {
    const [err, setErr] = useState<string | undefined>()
    const [errTxt, setErrTxt] = useState<string | undefined>()
    const [current, setCurrent] = useState<Additive>(initial)
    useEffect(() => {
        setCurrent(initial)
    }, [initial]);
    const handleFormChangeAmt = (val: number) => {
        const data = {...current};
        data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        const data = {...current};
        data.unit = val
        updateParent(data)
    }
    return <>
        <div className={"text-m"}>{current.additive}</div>
        {/* TODO: ensure floating */}
        <InputDecimal label={"Amount"} initial={initial.amount} updateParent={handleFormChangeAmt}  min={0.0} max={1000.0}/>{/* TODO: min/max ok?*/}

        {/*<NumericalAreaWithAbsolutes label="Amount" mode="floating" min={0.0} max={100.0} readonly={false}*/}
        {/*                            errorMessage={err} value={currentValue.amount.toString()}*/}
        {/*                            onChange={(val?: string) => {*/}
        {/*                                try {*/}
        {/*                                    const n = Number(val) // TODO: allow only numbers here*/}
        {/*                                    if (Number.isNaN(n)) {*/}
        {/*                                        setErr("NaN input")*/}
        {/*                                    } else {*/}
        {/*                                        val && handleFormChangeAmt(n)*/}
        {/*                                        setErr(undefined)*/}
        {/*                                    }*/}
        {/*                                } catch (e) {*/}
        {/*                                    setErr(JSON.stringify(e))*/}
        {/*                                }*/}
        {/*                            }}/>*/}
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt}
                                 value={current.unit.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function DowelEntryForNew({initial, updateParent}: {
    initial: DowelType,
    updateParent: (l: DowelType) => void
}) {
    // TODO: const [err, setErr] = useState<string | undefined>()
    const [errTxt, setErrTxt] = useState<string | undefined>()
    const [current, setCurrent] = useState<DowelType>(initial)
    useEffect(() => {
        setCurrent(initial)
    }, [initial]);
    // TODO: const [radiusDraft, setRadiusDraft] = useState(currentValue.size.toString())
    // TODO: useEffect(() => {
    //     setRadiusDraft(currentValue.size.toString())
    // }, [currentValue.size])
    const handleFormChangeRadius = (val: number) => {
        updateParent({...structuredClone(current), size: val}) // TODO: switch back if not work
        // const data = structuredClone(currentValue);
        // data.size = val
        // updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        updateParent({...structuredClone(current), units: val}) // TODO: switch back if not work
        // const data = structuredClone(currentValue);
        // data.units = val
        // updateParent(data)
    }
    return <>
        <div className={"text-m"}>{current.wood}</div>
        <InputDecimal initial={0.25} updateParent={handleFormChangeRadius} label={"Radius Magnitude"} min={0.0} max={1000.0}/>
        {/*<NumericalAreaWithAbsolutes label="Radius Magnitude" mode="floating" min={0.0} max={1000.0} readonly={false}*/}
        {/*                            errorMessage={err} value={radiusDraft}*/}
        {/*                            onChange={(val?: string) => {*/}
        {/*                                const next = val ?? ""*/}
        {/*                                setRadiusDraft(next) // TODO: do this on any other absolutes components...*/}
        {/*                                // allow in-progress values like "1."*/}
        {/*                                if (next === "" || next.endsWith(".")) {*/}
        {/*                                    setErr(undefined)*/}
        {/*                                    return*/}
        {/*                                }*/}
        {/*                                try {*/}
        {/*                                    const n = Number(val) // TODO: allow only numbers here*/}
        {/*                                    if (!Number.isNaN(n)) {*/}
        {/*                                        val && handleFormChangeRadius(n)*/}
        {/*                                        setErr(undefined)*/}
        {/*                                    } else {*/}
        {/*                                        setErr("NaN input")*/}
        {/*                                    }*/}
        {/*                                } catch (e) {*/}
        {/*                                    setErr(JSON.stringify(e))*/}
        {/*                                }*/}
        {/*                            }}/>*/}
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt}
                                 value={current.units.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function ParentDisplay(
    {parent, parentType}: {
        parent?: string,
        parentType?: string,
    }
) {
    if (parent == undefined) {
        return null
    }
    if (parentType === undefined) {
        return <div>
            {"Parent: "+parent+" (unknown type, should never happen)"/* TODO: THIS!*/}
        </div>
    }
    const txtFor = (typ: string, id: string) => {
        return typ + " " + id
    }
    const LinkFor = (typ: string, entryTyp: string, linkId: string, displayId: string) => {
        return <div>
            {"Parent: "}<EntryLinkIdWrapper props={{linkId: linkId, entryType: entryTyp}}>
                {txtFor(typ, displayId)}
            </EntryLinkIdWrapper>
        </div>
    }
    switch (parentType) {
        case 'fruit':
            return LinkFor("Fruit", parentType, parent, parent)
        case 'mss':
            return LinkFor("Spore Syringe", parentType, parent, parent)
        case 'plate':
            return LinkFor("Plate", parentType, parent, parent)
        case 'slant':
            return LinkFor("Slant", parentType, parent, parent)
        case 'stasisTube':
            return LinkFor("Stasis Tube", parentType, parent, parent)
        case 'jar':
            return LinkFor("Grain Jar", parentType, parent, parent)
        case 'lc':
            return LinkFor("Liquid Culture", parentType, parent, parent)
        case 'lcSyringe':
            return LinkFor("Liquid Culture Syringe", parentType, parent, parent)
        case 'bag':
            return LinkFor("Bag", parentType, parent, parent)
        case 'fruitingChamber':
            return LinkFor("Fruiting Chamber", parentType, parent, parent)
        case 'sporePrint':
            return LinkFor("Spore Print", parentType, parent, parent)
        case 'sporeSwab':
            return LinkFor("Spore Swab", parentType, parent, parent)
        default:
            return <div>{"Unknown parentType: " + parentType + " with ID " + parent}</div>
    }
}

// export function ProjectsArea({ // TODO: MOVE????????
//                                  projects, headerLevel, readonly, setProjects
//                              }: {
//                                  projects: string[],
//                                  setProjects?: (p?: string[]) => void
//                                  headerLevel?: number,
//                                  readonly?: boolean,
//                              }
// ) {
//     let ProjectsLabel = <div>{"Projects: "}</div>
//     // TODO: ADD TO A PROJECT????
//     // TODO: CREATE A PROJECT????
//     return <div>
//         {ProjectsLabel}
//         {projects.map((proj, i) => {
//             return <div key={i}>{proj}</div> // TODO: ENSURE OK
//         })}
//         {/* TODO: ADD A NEW PROJECT AREA */}
//         {/* TODO: Create a project link? */}
//     </div>
// }

export function GensInlineDisplay(
    {gensSinceSpore, gensSinceFruitOrSpore, dontDisplayGensFruitOrSpore, headerLevel, offset}: {
        gensSinceSpore?: number,
        gensSinceFruitOrSpore?: number,
        headerLevel?: number,
        dontDisplayGensFruitOrSpore?: boolean,
        offset?: number,
    }
) {
    return <div>
        <div>{"Generations since:"}</div>
        <div>
            <div>{"Spore: "}</div>
            <div>{gensSinceSpore || "unknown"}</div>
        </div>
        {(!dontDisplayGensFruitOrSpore) && <div>
            <div>{"Fruit or Spore: "}</div>
            <div>{gensSinceFruitOrSpore || "unknown"}</div>
        </div>}
    </div>
}

export function GensFormDisplay(
    {gensSinceSpore, gensSinceFruitOrSpore, dontDisplayGensFruitOrSpore}: {
        gensSinceSpore?: number,
        gensSinceFruitOrSpore?: number,
        dontDisplayGensFruitOrSpore?: boolean,

    }
) {
    return <>
        <div>{"Generations since:"}</div>
        <div>{"Spore: " + (gensSinceSpore || "unknown")}</div>
        {(!dontDisplayGensFruitOrSpore) && <div>{"Fruit or Spore: " + (gensSinceFruitOrSpore || "unknown")}</div>}
    </>
}

function picsKey(items: PicWithNotesIncoming[]): string {
    return items.map((p) =>
        [
            p.time,
            p.location || "",
            (p.notes || []).map((n) => `${n.time}:${n.note}`).join("^"),
        ].join("|")
    ).join("||");
}



export const PicsDisplay = (
    props: {
        pix?: PicWithNotesIncoming[],
        updateParent: (v: SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>) => void,
        readonly?: boolean,
        sectionHeader?: string,
        addButtonText?: string,
        headerLevel?: number,
        offset?: number,
    }
) => {
    const pwnfFor = (item: PicWithNotesIncoming): Data<PicWithNotesForm> => {
        return {data: {time: item.time, img: item.location, notes: InitialNotesState(item.notes)}, disabled: false}
    }
    const pwnfs = (items: PicWithNotesIncoming[]): Data<PicWithNotesForm>[] => {
        return items.map(v => {
            return pwnfFor(v)
        })
    }
    const pix = props.pix || [];
    const pixInitKey = picsKey(pix);

    const [existing, setExisting] = useState<Data<PicWithNotesForm>[]>(pwnfs(pix));
    const [created, setCreated] = useState<NewPicWithNotesForm[]>([]);

    useEffect(() => {
        setExisting(pwnfs(pix));
        setCreated([]);
    }, [pixInitKey]);

    const doUpdate = () => {
        props.updateParent({
            existing: existing,
            new: created,
        })
    }
    const updateExisting = (updated: Data<PicWithNotesForm>[]) => {
        setExisting(updated)
        doUpdate()
    }
    const updateNew = (updated: NewPicWithNotesForm[]) => {
        setCreated(updated)
        doUpdate()
    }
    const depth = useContext(DepthContext)
    // TODO: OVERHAUL WITH EITHER GRID OR FLEXBOX?
    return <div /*key={count}*/ className={"depthContainer depth" + depth}>
        <div className={"areaHeader"}>{props.sectionHeader || "Pictures"}</div>
        <div className={"picsGroup picsRows"}>{/* TODO: change to grid???*/}
            {pix.map((img, i) => {
                {/* TODO: REMOVE CURRENT FROM INPUTS! DO INITIAL INSTEAD!*/
                }
                return <PixRowExisting key={i} initial={img} readonly={props.readonly} updateParent={a => {
                    const upd = structuredClone(existing)
                    upd[i] = a
                    updateExisting(upd)
                }}/>
            })}
        </div>
        {!props.readonly && <PixRows initial={pix} addButtonText={props.addButtonText} updateParent={a => {
            const upd = structuredClone(a)
            updateNew(upd)
        }}/>}

    </div>
}

export const PixRowExisting = (
    {readonly, updateParent, initial}: {
        initial: PicWithNotesIncoming,
        readonly?: boolean,
        updateParent?: (d: Data<PicWithNotesForm>) => void
    }
) => {
    const pwnfFor = (item: PicWithNotesIncoming): Data<PicWithNotesForm> => {
        return {
            data: {
                time: item.time,
                img: item.location,
                notes: InitialNotesState(item.notes),
            }, disabled: false
        }
    }
    const [current, setCurrent] = useState<Data<PicWithNotesForm>>(pwnfFor(initial))
    useEffect(() => {
        setCurrent(pwnfFor(initial))// reset when initial changes
    }, [initial])
    const update = (updated: Data<PicWithNotesForm>) => {
        setCurrent(updated)
        updateParent && updateParent(updated)
    }
    const disabledClass = () => {
        return current.disabled ? " disabled" : ""
    }
    const leftArea = () => {
        return <div className={"picLeft" + disabledClass()}>
            {/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */}
            {/*<Image className={"picDisplay"} src={ImageLocationFor(initial.location)} alt={"existing image"}/>*/}
            <img className={"picDisplay"} src={ImageLocationFor(initial.location)} alt={"existing image"}/>
            {!readonly &&
                <button className={current.disabled ? "basicButtonSmall" : "removeButtonSmall"} onClick={(e) => {
                    e.stopPropagation();
                    const upd = structuredClone(current)
                    upd.disabled = !current.disabled
                    update(upd)
                }}>
                    {(current.disabled ? "Keep" : "Delete") + " this image"}{/* TODO: THIS IS NOT WORKING!!!!!*/}
                </button>}
        </div>
    }
    const rightArea = () => {
        return <div className={"picRight" + disabledClass()}>
            <DateArea readonly={true} when={initial.time}/>
            <NotesFormArea initial={initial.notes} readonly={readonly || false}
                           updateParent={(nts: AllEntries<Note>) => {
                               const updated = structuredClone(current)
                               updated.data.notes = nts
                               update(updated)
                           }} removeHeader={true}/>
        </div>
    }
    return <div className={"contentsOnly picRow" + disabledClass()}>
        {leftArea()}
        {rightArea()}
    </div>
}

export function MostRecentImageDisplay(
    {data, headerTxt, showHeader}: {
        data?: PicWithNotesIncoming,
        headerTxt?: string,
        showHeader?: boolean,
    }) {
    if (data === undefined) {
        return null
    }
    const mostRecentImageHeader = headerTxt || "Image Upload Date: "
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div className={"mriParent depthContainer depth" + depth}>
            <div>
                {/*<Image className={"picDisplay mri"} src={ImageLocationFor(data.location)} alt={"most recent image"}/>*/}
                <img className={"picDisplay mri"} src={ImageLocationFor(data.location)} alt={"most recent image"}/>
            </div>
            <div className={"mriInfoHolder"}>
                <DateArea pre={(showHeader ? mostRecentImageHeader : undefined)} when={data.time} readonly={true}/>
                <NotesAreaMostRecentImage readonly={true} current={{existing: dataFor(data.notes), new: []}}/>
            </div>
        </div>
    </DepthProvider>
}

// export const SpeciesArea = (
//     {
//         readonly, setSpecies, initial, headerLevel
//     }: {
//         readonly: boolean,
//         setSpecies?: (sp: SpeciesData | undefined) => void,
//         initial?: string,
//         headerLevel?: number,
//     }
// ) => {
//     let spArea: JSX.Element | null = null
//     if (readonly) {
//         spArea = <div>{"unknown"}</div>
//         if (initial !== undefined) {
//             spArea = <EntryLinkForId props={{
//                 linkId: initial.split(" ").join("_"), // TODO: FIX THIS! URLENCODE!
//                 displayId: initial,
//                 entryType: "species",
//                 openInNewTab: false, // TODO: ok?
//             }}/>
//         }
//     } else {
//         // TODO: CSS
//         spArea = <ExistingSpeciesSelector doSelect={(s) => {
//             setSpecies && setSpecies(s)
//         }}/>
//     }
//     return <div className={"areaWrapper"}>
//         <div className={"areaHeader"}>{"Species:"}</div>
//         <div>{spArea}</div>
//     </div>
// }

// export function SpeciesFormArea({species}:{
//     species?: string,
// }){
//     return <div>
//         {"Species: "+(species?species:"undefined")}{/* TODO: LINK!?*/}
//     </div>
// }
// export function SpeciesSubspeciesFormArea({species,subspecies}:{
//     species?: string,
//     subspecies?: string,
// }){
//     return <>
//         <SpeciesFormArea species={species}/>
//         {subspecies && <SubspeciesFormArea subspecies={subspecies}/>}
//     </>
// }

// export const SubspeciesArea = (
//     {
//         readonly, currentSpecies, initialSub, setSubspecies, headerLevel
//     }: {
//         readonly: boolean,
//         setSubspecies?: (sp: SubspeciesData | undefined) => void,
//         currentSpecies?: string,
//         initialSub?: string,
//         headerLevel?: number
//     }
// ) => {
//     if (currentSpecies === undefined) {
//         return null
//     }
//     let spArea: JSX.Element | null = null
//     if (readonly) {
//         if (initialSub !== undefined) {
//             spArea = <EntryLinkForId props={{
//                 displayId: initialSub,
//                 linkId: initialSub.split(" ").join("_"), // TODO: FIX! URLENCODE!
//                 entryType: "subspecies",
//                 openInNewTab: false, // TODO: ok?
//             }}/>
//         }
//     } else {
//         spArea = <ExistingSubSpeciesSelector species={currentSpecies} doSelect={(s) => {
//             setSubspecies && setSubspecies(s)
//         }}/>
//     }
//     return <div className={"areaWrapper"}>
//         <div className={"areaHeader"}>{"Subspecies: "}</div>
//         <div>{spArea}</div>
//     </div>
// }

// export const SalesArea = ( // TODO: MOVE???? // TODO: onClick??????
//     readonly: boolean,
//     initialSales: string[],
//     newSale?: () => void,
//     headerLevel?: number,
// ) => {
//     return <div>
//         <div>Sales</div>
//         {initialSales.map((sale) => {
//             const b58id = BinaryToBase58(sale)
//             return <div>
//                 <div>{"Sales"}</div> {/* TODO: MAKE THIS A LINK!!!!!*/}
//             </div>
//         })}
//         {!readonly && <button onClick={newSale}>{"New Sale"}</button>}
//     </div>
// }

export function SporePrintColorArea(
    {readonly, color, setColor}: {
        readonly: boolean,
        color?: string,
        setColor?: (s?: string) => void,
    }
) {
    return <NoSsr>
        <div className={"sporePrintColorArea"}>
            <div className={"areaHeader"}>{"Color: "}</div>
            {readonly ? <div>{color}</div> : <SporePrintColorSelector current={color} onSelect={setColor}/>}
        </div>
    </NoSsr>
}

export function SporePrintDensityArea(
    {readonly, density, setDensity}: {
        readonly: boolean,
        density?: string,
        setDensity?: (s?: string) => void,
    }
) {
    return <NoSsr>
        <div className={"sporePrintDensityArea"}>
            <div className={"areaHeader"}>{"Density: "}</div>
            {readonly ? <div>{density}</div> : <SporePrintDensitySelector current={density} onSelect={setDensity}/>}
        </div>
    </NoSsr>
}

export function StringOptionsSelector({queryKey, variant, current, onSelect}: {
    queryKey: string,
    variant: string,
    current?: string,
    onSelect?: (value?: string) => void,
}) {
    const {isPending, error, data} = useQuery({
        queryKey: [queryKey],
        queryFn: () => {
            return getOptionsResponse(variant)
        }
    })
    if (isPending || error !== null) {
        return <div>{isPending ? variant + " SELECTOR LOADING" : variant + " SELECTOR ERROR: " + error.message}</div>
    }
    return <SelectorFor disabled={onSelect === undefined} options={["", ...data]} initial={current || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect(undefined)
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}

export function SporePrintDensitySelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    return <StringOptionsSelector queryKey={"spore print densities"} variant={"sporePrintDensities"}
                                  current={current} onSelect={onSelect}/>
}

export function SporePrintColorSelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    return <StringOptionsSelector queryKey={"spore print colors"} variant={"sporePrintColors"} current={current}
                                  onSelect={onSelect}/>
}

export function DisposedDisplay(
    {readonly, initial, setDisposedOnParent}: {
        readonly: boolean,
        initial?: number, // TODO: switch these everywhere from disposed to initial.disposed
        setDisposedOnParent?: (n?: number) => void,
    }
) {
    const [dispose, setDispose] = useState(!!initial)
    // useEffect(() => {
    //     setDispose(!!disposed) // TODO: necessary? probably not
    // }, [disposed]);
    if (readonly || initial !== undefined) {
        return <NoSsr>
            <div className={"disposedArea"}>{/* TODO: reformat */}
                <div>{"Disposed: "+(initial ? NumberToDate(new Date(initial)) : "Not Yet Disposed")}</div>
            </div>
        </NoSsr>
    }
    const disposeOnParent = () => {
        const DisposalTime = Date.now()
        setDisposedOnParent && setDisposedOnParent(DisposalTime)
    }
    return <NoSsr>
        <div className={"disposedArea inlineChildren"}>
            <div>{"Disposed: "}</div>
            <input type={"checkbox"} checked={dispose} onClick={e => {
                e.stopPropagation()
                const doDispose = !dispose
                setDispose(doDispose)
                if (doDispose) {
                    disposeOnParent()
                } else {
                    setDisposedOnParent && setDisposedOnParent(undefined)
                }

            }}/>
        </div>
    </NoSsr>
}

export const ErrorDisplay = ({
                                 err, classNames
                             }: {
    err?: string
    classNames?: string
}) => {
    if (err === undefined) {
        return null
    }
    return <div className={"Error centerH" + (classNames ? " " + classNames : "")}>
        <p>{err}</p>
    </div>
}

export function StandardArea(
    {
        isStandard, setStandard, headerTxt, readonly, headerLevel
    }: {
        isStandard?: boolean,
        setStandard?: (is: boolean) => void
        headerTxt?: string
        readonly?: boolean,
        headerLevel?: number,
    }) {
    const stdTxt = headerTxt || "Standard: "
    const toggle = () => {
        setStandard && setStandard(!isStandard)
    }
    if (readonly) {
        return <div>{stdTxt + (isStandard ? "true" : "false")}</div>
    }
    return <div className={"inlineChildren mt-2"}>
        <div className={"mr-2"}>{stdTxt}</div>
        <input type="checkbox" checked={isStandard} onChange={toggle}/>
    </div>
}

export function NameArea(
    {
        currentName, setName, headerTxt, headerLevel, readonly, classNames, titleClasses
    }: {
        currentName?: string,
        setName?: (n: string) => void
        headerTxt?: string
        readonly?: boolean,
        headerLevel?: number,
        classNames?: string
        titleClasses?: string

    }) {
    if (readonly) {
        return <div>{headerTxt || "Name: "}{currentName || ""}</div>
    }
    return <div className={"my-5" + (classNames ? " " + classNames : "")}>
        <InlineTitle title={headerTxt || "Name: "} titleClasses={titleClasses}>
            <InputText value={currentName} readonly={false} errorMessage={"invalid name"} placeholder={"name"}
                       onChange={(s) => {
                           setName && setName(s || "")
                       }}/>
        </InlineTitle>
    </div>
}

export function InlineTitle(props: React.PropsWithChildren<{ title: string, titleClasses?: string }>) {
    return <>
        <div className={props.titleClasses || ""}>{props.title}</div>
        {props.children}
    </>
}

export function InputTextWithInlineTitle({currentContent, setContent, headerTxt, placeholder}: {
    currentContent?: string,
    setContent?: (n: string) => void
    headerTxt: string
    placeholder?: string
}) {
    return <div className={"my-5"}>
        <InputTextInlineTitle label={headerTxt} readonly={false} errorMessage={""} value={currentContent || ""}
                              placeholder={placeholder} onChange={(v) => {
            setContent && setContent(v || "")
        }}/>
    </div>
}

export function OpenMainPage(
    {
        type, txt, linkId, redirect // TODO: MAKE SURE USING B58IDs (or underlined) HERE EVERYWHERE
    }: {
        type: string
        redirect?: boolean
        linkId: string
        txt?: string
    }) {
    const isTopLevel = useContext(DepthContext) === 0
    const handleClick = (e: React.MouseEvent) => {
        e.preventDefault()
        const url = viewUrlFor(type, linkId)
        if (redirect) {
            window.location.assign(url)
            // redirect(url) // TODO: del if working
        } else {
            window.open(url, '_blank', 'noopener,noreferrer'); // TODO: ensure ok
        }

    }
    return isTopLevel ? null : <div className={"openMainPageButton"}>
        <button className={"basicButton"} onClick={(e) => {
            e.preventDefault() // Ensure other (parent) click handlers don't do anything
            handleClick(e)
        }}>{txt || "View Page"}</button>
    </div>
}

export function AliasesArea(
    {
        initial, readonly, updateParent
    }: {
        initial?: string[]
        readonly?: boolean
        updateParent?: (s: string[]) => void
    }) {
    const [existing, setExisting] = useState<Data<string>[]>(initial ? initial.map(v=>{return {data:v,disabled:false}}) : [])
    const [created, setCreated] = useState<Data<string>[]>([])
    const [reloadCount, setReloadCount] = useState(0)

    useEffect(() => {
        const ex = initial ? initial.map(v=>{return {data:v,disabled:false}}) : []
        setExisting(ex)
        setCreated([])
        setReloadCount(reloadCount+1)
        deliverUpdatesToParent({existing:ex,new:[]})
    }, [initial])
    const currentClone = ()=>{
        return {
            existing:structuredClone(existing),
            new: structuredClone(created)
        }
    }
    const deliverUpdatesToParent = (updated:AllEntries<string>) => {
        const v = structuredClone(updated)
        const existingToSend = v.existing.filter(v=>!v.disabled).map(v=>v.data)
        const newToSend = v.new.filter(v=>{return !v.disabled&&v.data!==""}).map(v=>v.data)
        updateParent && updateParent([...existingToSend, ...newToSend])
    }
    const updateExisting = (updated:Data<string>[])=>{
        setExisting(updated)
        const out = currentClone()
        out.existing = updated
        deliverUpdatesToParent(out)
    }
    const updateCreated = (updated:Data<string>[])=>{
        setCreated(updated)
        const out = currentClone()
        out.new = updated
        deliverUpdatesToParent(out)
    }
    if (readonly) {
        if (!initial) {
            return null
        }
        return <div>
            <div>{"Aliases :"}</div>
            {existing.map((a, i) => {
                return <div key={i+a.data}>{a.data}</div>
            })}
        </div>
    }
    const existingArea = () => {
        if (!initial || initial.length <= 0) {
            return null
        }
        return <>
            {existing.map((v,i)=>{
                return <div key={i} className={"existingAlias" + (v.disabled ? " disabled" : "")}>
                    <SingleAlias initial={initial[i]} readonly={readonly||false} updateParent={v=>{
                        const updated = structuredClone(existing)
                        updated[i] = structuredClone(v)
                        updateExisting(updated)
                    }} startEditing={false}/>
                    {readonly || <RemoveAliasButton disabled={v.disabled} click={() => {
                        const updated = structuredClone(existing)
                        updated[i].disabled = !v.disabled
                        updateExisting(updated)
                    }}/>}
                </div>
            })}
        </>
    }
    return <div>
        <div>{"Aliases :"}</div>
        {existingArea()}
        <NewAliasesSubArea count={reloadCount} updateParent={updateCreated} readonly={readonly||false}/>
    </div>
}

export function NewAliasesSubArea({count,readonly,updateParent}:{count:number,readonly:boolean,updateParent:(entries:Data<string>[]) => void}){
    const [aliases, setAliases] = useState<Data<string>[]>([])
    useEffect(() => {
        setAliases([]);
    }, [count]);
    if (readonly) {
        return null
    }
    const propagateUpdate = (updated:Data<string>[]) => {
        setAliases(updated)
        updateParent(structuredClone(updated).filter((item)=>{
            return !item.disabled && item.data!==""
        }))
    }
    const createNewAlias = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
        e.stopPropagation();
        // Do not update parent here. We don't want to propagate empty notes
        setAliases([...structuredClone(aliases), {data: "",disabled: false}])
    }
    return <div>
        {aliases.map((n, i) => {
            if (n.disabled) {
                return null
            }
            return <div key={i}>
                <SingleAlias readonly={false} startEditing={true} updateParent={nd => {
                    const updated = structuredClone(aliases)
                    updated[i].data = structuredClone(nd.data)
                    propagateUpdate(updated)
                }}/>
                <RemoveNewAliasButton click={() => {
                    const updated = structuredClone(aliases)
                    updated[i].disabled = true
                    propagateUpdate(updated)
                }}/>
            </div>
        })}
        <div>
            <button className={"basicButtonSmall"} onClick={createNewAlias}>{"Create New Alias"}</button>
        </div>
    </div>


}
function RemoveNewAliasButton({click}:{click:()=>void}){
    return <RemoveAliasButton disabled={false} click={click}/>
}
function RemoveAliasButton({disabled,click}:{disabled:boolean,click:()=>void}){
    return <RemoveToggle disabled={disabled} click={click} keptTxt={"Delete Alias"} removedTxt={"Don't Delete"} keptClass={"removeButtonSmall"} removedClass={"basicButtonSmall"}/>
}
export function SingleAlias(
    {
        initial,
        readonly,
        startEditing,
        updateParent,
    }: {
        initial?: string
        readonly:boolean
        startEditing?: boolean
        updateParent?: (n: Data<string>) => void
    }) {
    const [val, setVal] = useState<Data<string>>({data:initial||"",disabled:false})
    const [started, setStarted] = useState(false)
    const [editing, setEditing] = useState(startEditing ?? false)
    useEffect(()=>{
        setVal({data:initial||"",disabled:false})
        if (!started){
            setEditing(startEditing || false)
            setStarted(true)
        } else {
            setEditing(false)
        }
    },[initial])
    const handleChangeStr = (updated: Data<string>) => {
        setVal(updated)
        updateParent && updateParent(updated)
    }
    return <div className={"alias"}>
        {(!readonly && editing) ? <input name='txt' type="text" disabled={false}
                                         autoComplete="off" value={val.data}
                                         placeholder={"new alias"}
                                         className={"aliasValue rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
                                         onBlur={() => {
                                             setEditing(false)
                                         }}
                                         onChange={(e)=>{
                                             e.stopPropagation();
                                             const updated = structuredClone(val);
                                             updated.data = e.target.value
                                             handleChangeStr(updated)
                                         }}
        /> : <>
            <div>{val.data}</div><button className={"basicButtonSmall"} onClick={()=>{setEditing(!editing)}}>
            {"Edit Alias"}
        </button>
        </>}
    </div>
}
