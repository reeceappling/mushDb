'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoImportRequest,
    DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey,
    RequiredArrayOfType,
    RequiredKey,
} from "@/app/components/common";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl,
} from "@/app/components/accessControlClient";
import {EntryLinkWrapper, EntryLinkIdWrapper} from "@/app/components/formSubcomponents/entryLink";
import {DowelType, PlugsData} from "@/app/components/plugsServer";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {KnownFruitableArea} from "./formSubcomponents/knownFruitableArea";
import ReaderWriterSelector, {WriteRfidOvcArea} from "./formSubcomponents/readerWriterButtons/readerSelector";
import {AllEntries, OnViewCreatorQuadCol} from "./formSubcomponents/shared";
import {ACL} from "./accessControlServer";
import {ErrorDisplay, GensFormDisplay, ParentDisplay} from "./formSubcomponents/commonClient";
import ID from "./formSubcomponents/id";
import {PcRunArea, PcRunSelector} from "./pcRunClient";
import {InnocDisplay, TransfersOutDisplay} from "./transferClient";
import {SpeciesData} from "./speciesServer";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import { ExistingSpeciesSubspeciesSelector, SpeciesSubspeciesArea} from "./speciesClient";
import {WoodEntriesGroupForNew} from "@/app/components/formSubcomponents/plugs";
import {SalesArea} from "@/app/components/saleClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import TestAndValidate from "@/app/components/testing/untested";

export function AssertPlugs(input: any): asserts input is PlugsData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plugs assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['parentType', 'string'],
        ['parent', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['pcRun', 'string'],
        ['knownFruitable', 'boolean'],
        ['disposed', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plugs assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['dowelTypes', IsValidDowel]
    ])
    for (const [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('Plug assertion failure: required array key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('plugs assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['sales', (item) => {
            return typeof item === 'string'
        }],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plugs assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function IsValidDowel(input: any): boolean {
    try {
        AssertDowel(input)
        return true
    } catch (error) {
        console.error("dowel invalid")
        console.error(error)
        return false
    }
}

export function AssertDowel(input: any): asserts input is DowelType {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['wood', 'string'],
        ['size', 'number'],
        ['units', 'string'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Dowel assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    return
}

export default function PlugsDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<PlugsData>) {
    const [initial, setInitial] = useState(data)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(data.knownFruitable)
    const [pcRun, setPcRun] = useState<string | undefined>(data.pcRun)
    const [sales, setSales] = useState<string[] | undefined>(data.sales)
    const [disposed, setDisposed] = useState<number | undefined>(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    const [acl, setAcl] = useState<ACL>(initial.acl)
    // Helper states
    const [transfersOut, setTransfersOut] = useState<string[]>(data.transfersOut || [])
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: PlugsData) => {
        setInitial(updated)
        setPcRun(updated.pcRun)
        setKnownFruitable(updated.knownFruitable)
        setSales(updated.sales)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setTransfersOut(updated.transfersOut || [])
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const submit = () => {
        const body: any = {
            pcRun: pcRun, // optional. can only be set once
            knownFruitable: knownFruitable,
            disposed: disposed,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        DoUpdateRequest("plugs",initial._id, body, AssertPlugs, allCookies(cookies))
            .then(v=>{
                updateInitial(new PlugsData(v))
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=> {
        const disp = initial.disposed !== undefined
        return !disp ? [
            WriteRfidOvcArea(initial._id),
        ]:[]
    }
        // TODO: fruit?
        // TODO: create spore print
        // TODO: creat spore swab
    const isInnoculated = ()=>{
        return initial.species !== undefined
    }
    return (
        <DisplayFormWrapper entryType={"plugs"}>
            <ErrorDisplay err={err}/>
            <ID props={{id:data._id, txt:"Plugs Jar", entryType:"plugs", linkPage:false, allowOpenMainPage:false}}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                initialDisposed={initial.disposed}
                                                readonly={readonly} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    {isInnoculated()&&<SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>}
                    <PcRunArea binaryId={pcRun}/>
                </FlexedSinglesGroup>
                {isInnoculated()&&<FlexedSinglesGroup>
                    <InnocDisplay innoc={initial.innoc}/>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType}/>
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                </FlexedSinglesGroup>}
                {isInnoculated()&&<FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedSinglesGroup>}
            </FlexedArea>
            <div>
                <div className={"text-lg"}>{"Dowel Types"}</div>
                <DowelTypesTable data={initial.dowelTypes}/>
            </div>
            {isInnoculated()&&<SalesArea allowCreate={!readonly} sales={sales} readonly={readonly} setEntries={setSales}/>}
            {isInnoculated()&&<TransfersOutDisplay headerTxt={"Transfers"} thisId={initial._id} thisEntryType={"plugs"}
                                 transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly}/>}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>

            {readonly || <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

export function DowelTypesTable({data}: { data: DowelType[] }) {
    return <table>
        <tr>
            <th className={"mr-[2em]"}>{"Wood"}</th>
            <th className={"mr-[2em]"}>{"Radius"}</th>
        </tr>
        {data.map((item, i) => {
            return <tr key={item.wood + item.size + item.units + i}>
                <td className={"mr-[2em]"}>{item.wood}</td>
                <td className={"mr-[2em]"}>{item.size + " " + item.units}</td>
            </tr>
        })}
    </table>
}


export function PlugsImportDisplay({}: ImportDisplayInput) {
    const cookies = useContext(CookiesContext)
    const [dowelTypes, setDowelTypes] = useState<DowelType[]>([])
    const [gen, setGen] = useState<number | undefined>(undefined)
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const ImportEntry = () => {
        const body: any = {
            dowelTypes: dowelTypes,
            generation: gen,
            // optional
            species: species?._id, // Unused if non-inoculated
            subspecies: subspecies,
            knownFruitable: knownFruitable,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoImportRequest(body, "plugs", AssertPlugs, setErr, allCookies(cookies))
    }
    return <ImportEntryFormWrapper entryType={"plugs"}>

        {err != undefined && <div>{"Error: " + err}</div>}
        <div>
            <div className={"text-lg"}>{"Dowels: "}</div>
            <WoodEntriesGroupForNew currentEntries={dowelTypes} updateParent={setDowelTypes}/>
        </div>

        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {species && <>
            <GenerationInput updateParent={setGen}/>
            <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable}/>
        </>}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={ImportEntry}>{"Import Plugs"}</button>
    </ImportEntryFormWrapper>
}

export function NewPlugsForm(
    {handlers, pcRunIn}: { handlers: NewEntryInput<PlugsData>, pcRunIn?: PcRunData }
) {
    /* TODO: DOWEL TYPES AND AN OPTIONAL PC RUN FIELD! */
    const [dowelTypes, setDowelTypes] = useState<DowelType[]>([])
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const createPlugs = (e: React.MouseEvent) => {
        e.preventDefault()
        if (dowelTypes.length === 0) {
            setErr("Must have >0 dowel types")
            // TODO: validate dowel types
            return
        }
        for (let i = 0; i < dowelTypes.length; i++) {
            if (!dowelTypes[i] || dowelTypes[i].size <= 0 || dowelTypes[i].units === "") {
                setErr("Invalid dowels")
                return
            }
        }
        const body: any = {
            dowelTypes: dowelTypes,
            pcRun: pcRun?._id,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("plugs", body, AssertPlugs, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"plugs"}>
        <ErrorDisplay err={err}/>
        <div>
            <div className={"text-lg"}>{"Dowels: "}</div>
            <WoodEntriesGroupForNew currentEntries={dowelTypes} updateParent={setDowelTypes}/>
        </div>

        <div>
            <div>{"PC Run: "}</div>
            <div>
            {pcRunIn ? <EntryLinkIdWrapper props={{entryType: "pcRun", linkId: pcRunIn?._id, openInNewTab: true}}>
                    {pcRunIn._id}
                </EntryLinkIdWrapper>
                : <PcRunSelectorCloseable doSelect={setPcRun} txt={"PC Run: "} creatorInPage={handlers.isTopLevel} allowCreation={handlers.isTopLevel}/>
            }
            </div>
        </div>

        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton bottomButton"} onClick={createPlugs}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function PlugsListPageTable({data, onClick, withLink}: ListPageItems<PlugsData>) {
    let cols: ListTableColumn<PlugsData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Spec", (v) => v.species || "", true),
        NewColumn("Subspec", v => v.subspecies || "", true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: PlugsData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new PlugsData(v)}}/>
}

export function PlugsSelectorTable({data, onClick}: ListPageItems<PlugsData>) {
    return <PlugsListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function PlugsSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: PlugsData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: PlugsData[]): JSX.Element => {
        return <PlugsSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"plugs"} entryTypes={"plugs"} doSelect={doSelect} asserter={AssertPlugs}
                                   table={table}>
        {allowCreate && <NewPlugsForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
