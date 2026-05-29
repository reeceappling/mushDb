'use client'

import React, {JSX, useState} from "react";
import {
    DisplayFormWrapper,
    DisplayInput, ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    IsString,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn,
    NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ErrorDisplay,
    ParentDisplay,
    SpeciesArea,
    SubspeciesArea,
} from "@/app/components/formSubcomponents/commonClient";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    IsValidNote, NewEntryNotes,
    Note, NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import {BaseExternalUrl} from "@/app/components/Constants";
import DateArea from "@/app/components/formSubcomponents/date";
import {MssData} from "@/app/components/mssServer";
import ID from "@/app/components/formSubcomponents/id";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {TransfersOutDisplay} from "@/app/components/transferClient";
import {SaleArea} from "@/app/components/saleClient";
import {AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {SporePrintData, SporePrintSelectorCloseable} from "@/app/components/sporePrintServer";
import {WaterJarData, WaterJarSelectorCloseable} from "@/app/components/waterJarServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";


export function AssertMss(input: any): asserts input is MssData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['subspecies', 'string'],
        ['parent', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', IsString],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function MssImportDisplay({headerLevel}: ImportDisplayInput) { // TODO: USE ONLY FOR PURCHASED OR PRE-EXISTING MSS
    const [createdDate, setCreatedDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    // Non-required
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [notes, setNotes] = useState<Note[]>([])

    // const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [entriesCreated, setEntriesCreated] = useState<string[]>([])
    const [err, setErr] = useState<string | undefined>()
    const entriesCreatedDiv = ()=>{
        if(entriesCreated.length===0){
            return null
        }
        return <div>
            <div><div>{"Multispore syringes Created:"}</div></div>
            {entriesCreated.map((created,i)=>{
                const b58id = created
                return <EntryLink props={{displayedId:b58id, linkId: b58id, entryType:"mss", openInNewTab: false}}>{/* TODO: OPENINNEWTAB false ok? */}
                    <div key={i} className={i===entriesCreated.length-1?"lastCreated":"created"}>
                        {b58id}
                    </div>
                </EntryLink>
            })}
        </div>
    }
    const tryImport = (e: React.MouseEvent) => {
        e.preventDefault()
        if(species===undefined){
            setErr("Species field cannot be undefined")
            return
        }
        let body: any = {
            creationDate: createdDate,
            species: species._id, // TODO: validate on insert
        }
        subspecies && (body.subspecies = subspecies._id)
        notes.length>0 && (body.notes = notes)
        fetch(BaseExternalUrl+"/db/import/mss", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleTxtResponse) // TODO: change to json for reasons
            .then((id) => {
                setEntriesCreated([...entriesCreated, id])
                // TODO: REDIRECT!
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <ImportEntryFormWrapper entryType={"mss"}>
        <ErrorDisplay err={err}/>
        {entriesCreatedDiv()}
        <DateArea readonly={false} pre={"Created: "} when={Date.now()} updateParent={setCreatedDate}/>
        <SpeciesArea initial={species?._id} readonly={false} setSpecies={setSpecies}/>
        <SubspeciesArea initialSub={subspecies?._id} currentSpecies={species?._id} readonly={false} setSubspecies={setSubspecies}/>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton"} onClick={tryImport}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export default function MssDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertMss(data)
        const [initial, setInitial] = useState(data)

        const [sale, setSale] = useState(data.sale)
        const [disposed, setDisposed] = useState(data.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const updateInitial= (updated: MssData)=>{
            setInitial(updated)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setTransfersOut(updated.transfersOut || [])
        }
        const mssSubmit = () => {
            let body: any = {
                sale:sale,
                disposed:disposed,
                writeTagTo:writeTagTo,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            fetch(BaseExternalUrl+"/db/update/mss/"+data._id, {
                method: 'Post',
                body: JSON.stringify(body),
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertMss(entry)
                    updateInitial(entry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            WriteRfidOvcArea(initial._id),
        ]
        return <DisplayFormWrapper entryType={"mss"}>
            <ErrorDisplay err={err} headerLevel={headerLevel} />
            <ID id={data._id} txt={"Multispore Syringe"} entryType={"mss"} />
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel} canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <ParentDisplay parent={data.parent} parentType={"sporePrint"} headerLevel={headerLevel} />{/* TODO: can this be spore swab????*/}
                </FlexedSinglesGroup>
            </FlexedArea>
            <TransfersOutDisplay thisId={data._id} thisEntryType={"mss"} transfersOut={data.transfersOut} allowNewTransferCreation={!readonly} validTypesTo={["plate","slant","jar","bag"]} cookies={cookies}/>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes} />
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>/* TODO: ok?*/}
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                mssSubmit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: MuliSpore Syringe data format incorrect: " + err}</div>
    }
}

// this should only be used by the Spore Print Display Component
export function NewMssForm(
    {handlers,sporePrintIn,waterJarIn}: {handlers: NewEntryInput<MssData>, sporePrintIn?:SporePrintData, waterJarIn?:WaterJarData}){
    const [sporePrint, setSporePrint] = useState(sporePrintIn)
    const [waterJar, setWaterJar] = useState(waterJarIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (sporePrint === undefined){
            setErr("A sporePrint must be selected")
            return
        }
        if (waterJar === undefined){
            setErr("A waterJar must be selected")
            return
        }
        let body: any = {
            sporePrintId: sporePrint._id,
            waterJar: waterJar._id, // TODO: DO THIS ON THE GO SIDE!
            notes: notes,
            writeTagTo: writeTagTo,
        }

        fetch(BaseExternalUrl+"/db/create/mss", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertMss(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }

    return <NewEntryFormWrapper entryType={"mss"}>
        <ErrorDisplay err={err}/>
        { sporePrintIn === undefined && <SporePrintSelectorCloseable  onSelect={setSporePrint}/>}
        { waterJarIn === undefined && <WaterJarSelectorCloseable doSelect={setWaterJar} creatorInPage={false} allowCreation={false} />}
        <NewEntryNotes setNotes={setNotes} />
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function MssListPageTable({data, onClick, withLink}: ListPageItems<MssData>) {
    let cols: ListTableColumn<MssData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: MssData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"mss",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function MssSelectorTable({data, onClick}: ListPageItems<MssData>) {
    return <MssListPageTable data={data} onClick={onClick} withLink={true} />
}
export function MssSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: MssData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: MssData[]):JSX.Element=>{
        return <MssSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"mss"} entryTypes={"mss"} doSelect={doSelect} asserter={AssertMss}
                                   table={table}>
        {allowCreate && <NewMssForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}