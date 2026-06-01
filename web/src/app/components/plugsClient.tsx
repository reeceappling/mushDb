'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    clientPostRequestHeaders,
    createApiUrlFor, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse, importApiUrlFor,
    ImportDisplayInput, ImportEntryFormWrapper,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredArrayOfType,
    updateApiUrlFor,
    viewUrlFor,
} from "@/app/components/common";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth,} from "@/app/components/accessControlClient";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {DowelType, PlugsJar} from "@/app/components/plugsServer";
import {PcRunData} from "@/app/components/pcRunServer";
import {KnownFruitableArea} from "./formSubcomponents/knownFruitableArea";
import {ExistingSubSpeciesSelector} from "./subspeciesClient";
import ReaderWriterSelector from "./formSubcomponents/readerWriterButtons/readerSelector";
import {AllEntries} from "./formSubcomponents/shared";
import {ACL} from "./accessControlServer";
import {BaseExternalUrl} from "./Constants";
import {HandleErr} from "./userClient";
import {ErrorDisplay, GensFormDisplay, ParentDisplay} from "./formSubcomponents/commonClient";
import ID from "./formSubcomponents/id";
import {PcRunArea, PcRunSelector} from "./pcRunClient";
import {InnocDisplay, TransfersOutDisplay} from "./transferClient";
import {SpeciesData} from "./speciesServer";
import {SubspeciesData} from "./subspeciesServer";
import {redirect} from "next/navigation";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "./speciesClient";
import {WoodEntriesGroupForNew} from "@/app/components/formSubcomponents/plugs";
import {SalesArea} from "@/app/components/saleClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {AssertPlate} from "@/app/components/plateClient";
import {AssertJar} from "@/app/components/jarClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";

export function AssertPlugs(input: any): asserts input is PlugsJar {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plugs assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
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
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plugs assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['dowelTypes', IsValidDowel]
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('Plug assertion failure: required array key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plug assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['sales', (item) => {
            return typeof item === 'string'
        }],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plugs assertion failure: optional array key ' + key + ' was not valid');
        }
    }
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
    let requiredSimpleKeys = new Map<string, string>([
        ['wood', 'string'],
        ['size', 'number'],
        ['units', 'string'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Dowel assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    return
}

export default function PlugsDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    const [initial, setInitial] = useState(data as PlugsJar)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(data.knownFruitable)
    const [pcRun, setPcRun] = useState<string | undefined>(data.pcRun)
    const [sales, setSales] = useState<string[] | undefined>(data.sales)
    const [disposed, setDisposed] = useState<number | undefined>(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
    // Helper states
    const [transfersOut, setTransfersOut] = useState<string[]>(data.transfersOut || [])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: PlugsJar) => {
        setInitial(updated)
        setPcRun(updated.pcRun)
        setKnownFruitable(updated.knownFruitable)
        setSales(updated.sales)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setTransfersOut(updated.transfersOut || [])
        setAcl(updated.acl)
    }
    const submit = () => {
        let body: any = {
            pcRun: pcRun,
            knownFruitable: knownFruitable,
            disposed: disposed,
            notes: notes,
            writeTagTo: writeTagTo,
            acl: MarshalAcl(acl),
        }
        // TODO: do we want pics on this?
        DoUpdateRequest("plugs",initial._id, body, AssertPlugs)
            .then(updateInitial)
            .catch(ErrHandler(setErr))

        // fetch(updateApiUrlFor("plugs", data._id), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body)
        // }).then(HandleJsonResponse)
        //     .then((entry) => {
        //         AssertPlugs(entry)
        //         updateInitial(entry)
        //     })
        //     .catch((err) => {
        //         HandleErr(err, setErr)
        //     });
    }
    return (
        <DisplayFormWrapper entryType={"plugs"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID txt={"Plugs Jar"} id={initial._id} entryType={"plugs"} linkPage={false}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                disposed={disposed}
                                                readonly={readonly} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>
                    <PcRunArea binaryId={pcRun}/> {/* TODO: ENSURE OK AND ALLOWS USER TO INPUT*/}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <InnocDisplay innoc={initial.innoc}/>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <div>
                <div className={"text-lg"}>{"Dowel Types"}</div>
                <DowelTypesTable data={initial.dowelTypes}/>
            </div>
            <SalesArea allowCreate={!readonly} sales={sales} readonly={readonly} setEntries={setSales}/>
            <TransfersOutDisplay headerTxt={"Transfers"} thisId={initial._id} thisEntryType={"plugs"}
                                 transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly} cookies={cookies}/>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>

            {readonly || <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>}
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


export function PlugsImportDisplay({cookies}: ImportDisplayInput) {
    const [dowelTypes, setDowelTypes] = useState<DowelType[]>([])
    const [gen, setGen] = useState<number | undefined>(undefined)
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const ImportEntry = () => {
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        let body: any = {
            dowelTypes: dowelTypes,
            gen: gen,
            species: species._id,
            subspecies: subspecies?._id,
            knownFruitable: knownFruitable,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        fetch(importApiUrlFor("plugs"), {
            method: "POST",
            headers: clientPostRequestHeaders,
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((newItem) => {
                AssertPlugs(newItem)
                redirect(viewUrlFor("plugs", newItem._id))
            })
            .catch(ErrHandler(setErr));
    }
    return <ImportEntryFormWrapper entryType={"plugs"}>

        {err != undefined && <div>{"Error: " + err}</div>}
        <div>
            <div className={"text-lg"}>{"Dowels: "}</div>
            <WoodEntriesGroupForNew currentEntries={dowelTypes} updateParent={setDowelTypes}/>
        </div>
        <GenerationInput updateParent={setGen}/>
        <div className={"centerH"}>
            <ExistingSpeciesSelector initialSpecies={species?._id}
                                     doSelect={(spec?: SpeciesData) => {
                                         setSpecies(spec)
                                         setSubspecies(undefined)
                                     }}/>
        </div>
        {species !== undefined ? <div className={"centerH"}>
            <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        </div> : null}
        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={ImportEntry}>{"Import Plugs"}</button>
    </ImportEntryFormWrapper>
}

export function NewPlugsForm(
    {handlers, pcRunIn}: { handlers: NewEntryInput<PlugsJar>, pcRunIn?: PcRunData }
) {
    /* TODO: DOWEL TYPES AND AN OPTIONAL PC RUN FIELD! */
    const [dowelTypes, setDowelTypes] = useState<DowelType[]>([])
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const errHandler = ErrHandler(setErr)
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
        let body: any = {
            dowelTypes: dowelTypes,
            pcRun: pcRun?._id,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("plugs", body, AssertPlugs)
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("plugs"), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body)
        // })
        //     .then(HandleJsonResponse)
        //     .then((entry) => {
        //         AssertPlugs(entry)
        //         handlers.onCreate && handlers.onCreate(entry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    return <NewEntryFormWrapper entryType={"plugs"}>
        <ErrorDisplay err={err}/>
        <div>
            <div className={"text-lg"}>{"Dowels: "}</div>
            <WoodEntriesGroupForNew currentEntries={dowelTypes} updateParent={setDowelTypes}/>
        </div>


        {pcRunIn ? <div>{"PC Run: "}
                <EntryLinkWrapper props={{entryType: "pcRun", linkId: pcRunIn?._id, openInNewTab: true}}>
                    {pcRunIn._id}
                </EntryLinkWrapper>
            </div>
            : <PcRunSelector doSelect={setPcRun} allowCreate={true}/>
        }
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={createPlugs}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function PlugsListPageTable({data, onClick, withLink}: ListPageItems<PlugsJar>) {
    let cols: ListTableColumn<PlugsJar>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: PlugsJar) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "plugs", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function PlugsSelectorTable({data, onClick}: ListPageItems<PlugsJar>) {
    return <PlugsListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function PlugsSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: PlugsJar | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: PlugsJar[]): JSX.Element => {
        return <PlugsSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"plugs"} entryTypes={"plugs"} doSelect={doSelect} asserter={AssertPlugs}
                                   table={table}>
        {allowCreate && <NewPlugsForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
