'use client'

import {JSX, useContext, useEffect, useState} from "react";
import {Liquid} from "./liquids";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {AllEntries, Data, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {
    ImageLocationFor,
    NewPicWithNotesForm,
    PicWithNotesForm,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {PixRowNew} from "@/app/components/formSubcomponents/commonClient2";
import {
    InputText,
    InputTextInlineTitle,
    InputTextWithSmallTitle,
    NumericalArea,
    NumericalAreaWithAbsolutes
} from "./numericInput";
import DateArea, {NumberToDate} from "./date";
import NotesArea, {NotesAreaOld, Note, NotesAreaMostRecentImage, NotesGrid} from "./notes";
import {SpeciesData} from "@/app/components/speciesServer";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {ExistingSubSpeciesSelector, SubspeciesFormArea} from "@/app/components/subspeciesClient";
import {NoSsr} from "@mui/material";
import {useQuery} from "@tanstack/react-query";
import {BaseExternalUrl} from "@/app/components/Constants";
import {HandleJsonResponse} from "@/app/components/common";
import {SelectorFor} from "@/app/components/selector";
import {redirect} from "next/navigation";
import TextBoxArea from "@/app/components/formSubcomponents/singleTextBoxArea";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import TextBox from "@/app/components/formSubcomponents/textbox";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {DepthContext, DepthProvider} from "./depthContext/depth";
import {dataFor} from "@/app/components/agarRecipeClient";
import TestAndValidate from "@/app/components/testing/untested";

// export function OnClickWrapper(props: React.PropsWithChildren<{ handleClick?: () => void }>) {
//     return <div className={"hoverClickable"} onClick={(e) => {
//         e.stopPropagation() // TODO: ok?
//         e.preventDefault() // TODO: ok?
//         props.handleClick && props.handleClick()
//     }}>{props.children}</div>
// }


export function LiquidEntryForNew({currentValue, updateParent, key}: {
    currentValue: Liquid,
    updateParent: (l: Liquid) => void
    key: string
}) {
    const [err, setErr] = useState<string | undefined>()
    const handleFormChangePct = (val: number) => {
        let data = {...currentValue};
        data.pct = val
        updateParent(data)
    }
    return <>
        <div key={key+"title"} className={"text-m"}>{currentValue.name}</div>
        <NumericalAreaWithAbsolutes key={key+"inp"} label="Percentage by volume" mode="floating" min={0.0} max={1.0} readonly={false}
                                    errorMessage={err} value={currentValue.pct.toString()} onChange={(val?: string) => {
            try {
                const n = Number(val) // TODO: allow only numbers here
                if (Number.isNaN(n)) {
                    setErr("NaN input")
                } else {
                    val && handleFormChangePct(n)
                    setErr(undefined)
                }
            } catch (e) {
                setErr(JSON.stringify(e))
            }
        }}/>
    </>
}

export function NutrientEntryForNew({currentValue, updateParent, key}: {
    currentValue: Nutrient,
    updateParent: (l: Nutrient) => void
    key: string
}) {
    const [err, setErr] = useState<string | undefined>()
    const [errTxt, setErrTxt] = useState<string | undefined>()
    const handleFormChangeAmt = (val: number) => {
        let data = {...currentValue};
        data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        let data = {...currentValue};
        data.unit = val
        updateParent(data)
    }
    return <>
        <div key={key+"title"} className={"text-m"}>{currentValue.nutrient}</div>
        <NumericalAreaWithAbsolutes key={key+"inp"} label="Amount" mode="floating" min={0.0} max={1.0} readonly={false}
                                    errorMessage={err} value={currentValue.amount.toString()} onChange={(val?: string) => {
            try {
                const n = Number(val) // TODO: allow only numbers here
                if (Number.isNaN(n)) {
                    setErr("NaN input")
                } else {
                    val && handleFormChangeAmt(n)
                    setErr(undefined)
                }
            } catch (e) {
                setErr(JSON.stringify(e))
            }
        }}/>
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt} value={currentValue.amount.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function SugarEntryForNew({currentValue, updateParent, key}: {
    currentValue: Sugar,
    updateParent: (l: Sugar) => void
    key: string
}) {
    const [err, setErr] = useState<string | undefined>()
    const [errTxt, setErrTxt] = useState<string | undefined>()
    const handleFormChangeAmt = (val: number) => {
        let data = {...currentValue};
        data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        let data = {...currentValue};
        data.unit = val
        updateParent(data)
    }
    return <>
        <div key={key+"title"} className={"text-m"}>{currentValue.type}</div>
        <NumericalAreaWithAbsolutes key={key+"inp"} label="Amount" mode="floating" min={0.0} max={1.0} readonly={false}
                                    errorMessage={err} value={currentValue.amount.toString()} onChange={(val?: string) => {
            try {
                const n = Number(val) // TODO: allow only numbers here
                if (Number.isNaN(n)) {
                    setErr("NaN input")
                } else {
                    val && handleFormChangeAmt(n)
                    setErr(undefined)
                }
            } catch (e) {
                setErr(JSON.stringify(e))
            }
        }}/>
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt} value={currentValue.amount.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function AdditiveEntryForNew({currentValue, updateParent, key}: {
    currentValue: Additive,
    updateParent: (l: Additive) => void
    key: string
}) {
    const [err, setErr] = useState<string | undefined>()
    const [errTxt, setErrTxt] = useState<string | undefined>()
    const handleFormChangeAmt = (val: number) => {
        let data = {...currentValue};
        data.amount = val
        updateParent(data)
    }
    const handleFormChangeUnit = (val: string) => {
        let data = {...currentValue};
        data.unit = val
        updateParent(data)
    }
    return <>
        <div key={key+"title"} className={"text-m"}>{currentValue.additive}</div>
        <NumericalAreaWithAbsolutes key={key+"inp"} label="Amount" mode="floating" min={0.0} max={1.0} readonly={false}
                                    errorMessage={err} value={currentValue.amount.toString()} onChange={(val?: string) => {
            try {
                const n = Number(val) // TODO: allow only numbers here
                if (Number.isNaN(n)) {
                    setErr("NaN input")
                } else {
                    val && handleFormChangeAmt(n)
                    setErr(undefined)
                }
            } catch (e) {
                setErr(JSON.stringify(e))
            }
        }}/>
        <InputTextWithSmallTitle label="Unit" readonly={false} errorMessage={errTxt} value={currentValue.amount.toString()} onChange={(val?: string) => {
            try {
                val && handleFormChangeUnit(val)
            } catch (e) {
                setErrTxt(JSON.stringify(e))
            }
        }}/>
    </>
}

export function ParentDisplay(
    {parent, parentType, headerLevel}: {
        parent?: string,
        parentType?: string,
        headerLevel?: number,
    }
) {
    if (parent == undefined) {
        return null
    }
    if (parentType === undefined) {
        return <div>{"Error: PARENT TYPE UNDEFINED"}</div>
    }
    const txtFor = (typ:string,id:string)=>{
        return typ+" "+id
    }
    const LinkFor = (typ:string,entryTyp:string,linkId:string, displayId:string)=>{
        return <div>
            {"Parent: "}<EntryLinkWrapper props={{linkId:linkId,entryType:entryTyp}}>{txtFor(typ, displayId)}</EntryLinkWrapper>
        </div>
    }
    switch (parentType) {
        case 'fruit':
            return LinkFor("Fruit",parentType,parent,parent)
        case 'mss':
            return LinkFor("Spore Syringe",parentType,parent,parent)
        case 'plate':
            return LinkFor("Plate",parentType,parent,parent)
        case 'slant':
            return LinkFor("Slant",parentType,parent,parent)
        case 'stasisTube':
            return LinkFor("Stasis Tube",parentType,parent,parent)
        case 'jar':
            return LinkFor("Grain Jar",parentType,parent,parent)
        case 'lc':
            return LinkFor("Liquid Culture",parentType,parent,parent)
        case 'lcSyringe':
            return LinkFor("Liquid Culture Syringe",parentType,parent,parent)
        case 'bag':
            return LinkFor("Bag",parentType,parent,parent)
        case 'fruitingChamber':
            return LinkFor("Fruiting Chamber",parentType,parent,parent)
        case 'sporePrint':
            return LinkFor("Spore Print",parentType,parent,parent)
        case 'sporeSwab':
            return LinkFor("Spore Swab",parentType,parent,parent)
        default:
            return <div>{"Unknown parentType: " + parentType + " with ID "+parent}</div>
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
            <div >{gensSinceSpore || "unknown"}</div>
        </div>
        {(!dontDisplayGensFruitOrSpore) && <div>
            <div>{"Fruit or Spore: "}</div>
            <div>{gensSinceFruitOrSpore || "unknown"}</div>
        </div>}
    </div>
}
export function GensFormDisplay(
    {gensSinceSpore, gensSinceFruitOrSpore, dontDisplayGensFruitOrSpore, headerLevel, offset}: {
        gensSinceSpore?: number,
        gensSinceFruitOrSpore?: number,
        headerLevel?: number,
        dontDisplayGensFruitOrSpore?: boolean,
        offset?: number,
    }
) {
    return <>
        <div>{"Generations since:"}</div>
        <div>{"Spore: "+(gensSinceSpore || "unknown")}</div>
        {(!dontDisplayGensFruitOrSpore) && <div>{"Fruit or Spore: "+(gensSinceFruitOrSpore || "unknown")}</div>}
    </>
}

export const PicsDisplay = (
    props: {
        pix?: SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>,
        updateParent: (v: SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>) => void,
        readonly?: boolean,
        sectionHeader?: string,
        addButtonText?: string,
        headerLevel?: number,
        offset?: number,
    }
) => {
    // TODO: fix
    const addNewImageFields = () => { // TODO: formatting on addNewImageFields is incorrect
        // TODO: unsure if needed: e.preventDefault() (e: MouseEvent)
        let data = {...(props.pix || {existing: [], new: []})}
        data.new = [...(props.pix ? props.pix.new : []), {
            data: {time: Date.now(), img: undefined, notes: {existing: [], new: []}},
            disabled: false
        }]
        props.updateParent(data)
    }
    const depth = useContext(DepthContext)
    // TODO: OVERHAUL WITH EITHER GRID OR FLEXBOX?
    return <div /*key={count}*/ className={"depthContainer depth"+depth}>
        <div className={"areaHeader"}>{props.sectionHeader || "Pictures"}</div>
        {/* TODO: change?*/}
        <div className={"picsGroup picsRows"}>{/* TODO: change to grid???*/}
            {props.pix?.existing.map((img, i) => {
                return <PixRowExisting key={i} readonly={props.readonly} updateParent={a => {
                    let upd = {...(props.pix || {existing: [], new: []})}
                    upd.existing[i] = a
                    props.updateParent(upd)
                }
                } current={img}/>
            })}
            { /* Confirmed working! TODO EXCEPT NOTES */}
            {!props.readonly && props.pix?.new.map((img, i) => { // TODO: FIX IMAGE
                if (img.disabled) {
                    return null
                }
                return <PixRowNew key={i} current={img.data} updateParent={a => {
                    let upd = {...(props.pix || {existing: [], new: []})}
                    upd.new[i] = {data: a, disabled: false}
                    props.updateParent(upd)
                }} remv={() => {
                    // TODO: ensure correct!
                    let upd = {...(props.pix || {existing: [], new: []})}
                    upd.new[i].disabled = true
                    let toParent = {...upd}
                    toParent.new = toParent.new.filter(v => !v.disabled)
                    props.updateParent(toParent)
                }}/>
            })}
        </div>
        <div className={"centerH"}>
            <button className={"greenButton"} onClick={addNewImageFields}>{"Add new image"}</button>
        </div>

    </div>
}

export const PixRowExisting = (
    {readonly, updateParent, current}: {
        current: Data<PicWithNotesForm>,
        readonly?: boolean,
        updateParent?: (d: Data<PicWithNotesForm>) => void
    }
) => {
    const leftArea = () => {
        return <div className={"picLeft" + (current.disabled ? " disabled" : "")}>
            {/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */}
            <img className={"picDisplay"} src={ImageLocationFor(current.data.img)} alt={"existing image"}/>
            {!readonly && <button className={current.disabled?"basicButtonSmall":"removeButtonSmall"} onClick={() => {
                let upd = {...current}
                upd.disabled = !current.disabled
                updateParent && updateParent(upd)
            }}>
                {(current.disabled ? "ENABLE" : "DISABLE") + " THIS IMAGE"}
            </button>}
        </div>
    }
    const rightArea = () => {
        return <div className={"picRight" + (current.disabled ? " disabled" : "")}>
            <DateArea readonly={true} when={current.data.time}/> {/* TODO: INITIAL OR CURRENT?*/}
            {/* TODO: try with NotesGrid instead? */}
            <NotesGrid readonly={readonly} current={current.data.notes}
                          updateParent={(nts: AllEntries<Note>) => {
                           let out = {...current}
                           out.data.notes = nts
                           updateParent && updateParent(out)
                       }}/>
            {/*<NotesAreaOld readonly={readonly} current={current.data.notes}*/}
            {/*              updateParent={(nts: AllEntries<Note>) => {*/}
            {/*               let out = {...current}*/}
            {/*               out.data.notes = nts*/}
            {/*               updateParent && updateParent(out)*/}
            {/*           }}/>*/}
        </div>
    }
    return <div className={"contentsOnly picRow " + (current.disabled ? "disabled" : "") + ""}>
        {leftArea()}
        {rightArea()}
    </div>
}

// TODO: NEEDS MAJOR FIX!!!

export function MostRecentImageDisplay(
    {data, headerLevel, headerTxt, showHeader}: {
        data?: PicWithNotesIncoming, // TODO: change to File or string?????
        headerTxt?: string,
        headerLevel?: number,
        showHeader?: boolean,
    }) {
    if (data === undefined) {
        return null
    }
    const mostRecentImageHeader = headerTxt || "Image Upload Date: "
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <TestAndValidate todos={["reformat. Notes should be to the right"]}>
        <div className={"mriParent depthContainer depth"+depth}>
            <div>
                <img className={"picDisplay mri"} src={ImageLocationFor(data.location)} alt={"most recent image"}/>
            </div>
            <div className={"mriInfoHolder"}>
            {/*<div className={"mriInfoHolder"}>*/}
                <DateArea pre={(showHeader ? mostRecentImageHeader : undefined)} when={data.time} readonly={true}/>
                <NotesAreaMostRecentImage readonly={true} current={{existing: dataFor(data.notes), new: []}}/>
            </div>
        </div>
        </TestAndValidate>
    </DepthProvider>
}

export const SpeciesArea = (
    {
        readonly, setSpecies, initial, headerLevel
    }: {
        readonly: boolean,
        setSpecies?: (sp: SpeciesData | undefined) => void,
        initial?: string,
        headerLevel?: number,
    }
) => {
    let spArea: JSX.Element | null = null
    if (readonly) {
        spArea = <div>{"unknown"}</div>
        if (initial !== undefined) {
            spArea = <EntryLink props={{
                displayedId: initial,
                linkId: initial.split(" ").join("_"),
                entryType: "species"
            }}>{initial}</EntryLink>
        }
    } else {
        // TODO: CSS
        spArea = <ExistingSpeciesSelector doSelect={(s) => {
            setSpecies && setSpecies(s)
        }}/>
    }
    return <div className={"areaWrapper"}>
        <div className={"areaHeader"}>{"Species:"}</div>
        <div>{spArea}</div>
    </div>
}

export function SpeciesFormArea({species}:{
    species?: string,
}){
    return <div>
        {"Species: "+(species?species:"undefined")}{/* TODO: LINK!?*/}
    </div>
}
export function SpeciesSubspeciesFormArea({species,subspecies}:{
    species?: string,
    subspecies?: string,
}){
    return <>
        <SpeciesFormArea species={species}/>
        {subspecies && <SubspeciesFormArea subspecies={subspecies}/>}
    </>
}

export const SubspeciesArea = (
    {
        readonly, currentSpecies, initialSub, setSubspecies, headerLevel
    }: {
        readonly: boolean,
        setSubspecies?: (sp: SubspeciesData | undefined) => void,
        currentSpecies?: string,
        initialSub?: string,
        headerLevel?: number
    }
) => {
    if (currentSpecies === undefined) {
        return null
    }
    let spArea: JSX.Element | null = null
    if (readonly) {
        if (initialSub !== undefined) {
            spArea = <EntryLink props={{
                displayedId: initialSub,
                linkId: initialSub.split(" ").join("_"),
                entryType: "subspecies"
            }}>{initialSub}</EntryLink> // TODO: ensure subspecies is correct entryType
        }
    } else {
        spArea = <ExistingSubSpeciesSelector species={currentSpecies} doSelect={(s) => {
            setSubspecies && setSubspecies(s)
        }}/>
    }
    return <div className={"areaWrapper"}>
        <div className={"areaHeader"}>{"Subspecies: "}</div>
        <div>{spArea}</div>
    </div>
}

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

export function StringOptionsSelector({queryKey, selectorFor, current, onSelect}: {
    queryKey: string,
    selectorFor: string,
    current?: string,
    onSelect?: (value?: string) => void,
}) {
    const {isPending, error, data} = useQuery({
        queryKey: [queryKey],
        queryFn: () => {
            // TODO: delete lines before fetch for the real server
            const map = new Map<string, string>();
            map.set("fixme1-" + queryKey, "outgrew plate");
            map.set("fixme2-dens" + queryKey, "parent was contaminated");
            map.set("fixme3-dens" + queryKey, "transferring a specific sector");
            return map;
            // TODO: reenable
            fetch(BaseExternalUrl + "/options/" + queryKey).then(HandleJsonResponse).then((resJson) => {
                return ConvertObjectToStringMap(resJson)
            }).catch((e) => {
                throw e
            })
        },
    })
    if (isPending || error !== null) {
        return <div>{isPending ? selectorFor + " SELECTOR LOADING" : selectorFor + " SELECTOR ERROR: " + error.message}</div>
    }
    // TODO: maybe do selector a little differently because this one is a map?
    return <SelectorFor disabled={onSelect === undefined} options={["", ...data.keys()]} initial={current || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect()
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}

export function ConvertObjectToStringMap(obj: { [key: string]: string }): Map<string, string> {
    const map = new Map<string, any>();
    for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
            map.set(key, obj[key]);
        }
    }
    return map;
}

export function SporePrintDensitySelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    return <StringOptionsSelector queryKey={"sporePrintDensities"} selectorFor={"spore print densities"}
                                  current={current} onSelect={onSelect}/>
}

export function SporePrintColorSelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    return <StringOptionsSelector queryKey={"sporePrintColors"} selectorFor={"spore print colors"} current={current}
                                  onSelect={onSelect}/>
}

export function DisposedDisplay(
    {readonly, disposed, setDisposedOnParent}: {
        readonly: boolean,
        disposed?: number,
        setDisposedOnParent?: (n?: number) => void,
    }
) {
    if (readonly || (disposed !== undefined)) {
        return <NoSsr>
            <div className={"disposedArea"}>
                <div>{disposed ? "Disposed: " : "Not Yet Disposed"}</div>
                {disposed && <div>{NumberToDate(new Date(disposed))}</div>}
            </div>
        </NoSsr>
    }
    const dispose = () => {
        let DisposalTime = Date.now()
        setDisposedOnParent && setDisposedOnParent(DisposalTime)
    }
    return <NoSsr>
        <div className={"disposedArea"}>
            <div className={"areaHeader"}>{"Not Yet Disposed: "}</div>
            <button className={"removeButton"} onClick={dispose}>{"Dispose"}</button>
        </div>
    </NoSsr>
}

export const ErrorDisplay = ({ // TODO: USE
                                 err, headerLevel, offset
                             }: {
    err?: string
    headerLevel?: number // TODO: REMOVE
    offset?: number // TODO: REMOVE?
}) => {
    if (err === undefined) {
        return null
    }
    return <div className={"Error centerH"}>
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
    return <div className={"my-5"+(classNames?" "+classNames:"")}>
        <InlineTitle title={headerTxt || "Name: "} titleClasses={titleClasses}>
            <InputText value={currentName} readonly={false} errorMessage={"invalid name"} placeholder={"name"} onChange={(s)=>{setName && setName(s||"")}}  />
        </InlineTitle>
    </div>
}

export function InlineTitle(props:React.PropsWithChildren<{title:string,titleClasses?:string}>){
    return <>
    <div className={props.titleClasses||""}>{props.title}</div>
        {props.children}
    </>
}

export function InputTextWithInlineTitle({currentContent, setContent, headerTxt, placeholder}:{
    currentContent?: string,
    setContent?: (n: string) => void
    headerTxt: string
    placeholder?: string
}){
    return <div className={"my-5"}>
        <InputTextInlineTitle label={headerTxt} readonly={false} errorMessage={""} value={currentContent || ""} placeholder={placeholder} onChange={(v)=>{setContent && setContent(v || "")}}/>
    </div>
}

export function OpenMainPage(
    {
        type, txt, linkId // TODO: MAKE SURE USING B58IDs (or underlined) HERE EVERYWHERE
    }: {
        type: string
        redirect?: boolean
        linkId: string
        txt?: string
    }) {
    const isTopLevel = useContext(DepthContext) === 0
    const handleClick = (e: React.MouseEvent) => {
        e.preventDefault()
        const url = BaseExternalUrl + "/view/" + type + "/" + linkId
        if (redirect) {
            redirect(url)
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
        aliases, readonly, headerLevel, offset, updateParent
    }: {
        aliases?: string[]
        readonly?: boolean
        headerLevel?: number
        offset?: number
        updateParent?: (s: string[]) => void
    }) {
    if (readonly) {
        return <div>
            <div>{"Aliases :"}</div>
            {(aliases || []).map((a, i) => {
                return <div key={i}>{a}</div>
            })}
        </div>
    }
    return <div>
        <div>{"Aliases :"}</div>
        {/* TODO: DONT USE TEXTBOXAREA */}
        <TextBoxArea readonly={false} initialValues={(aliases || []).map((((a: string) => {
            return {data: a, disabled: false}
        })))} updateParent={(v) => {
            let newVals = v.new.map((n) => {
                return n.data
            })
            updateParent && updateParent(newVals)
        }}/>
    </div>
}