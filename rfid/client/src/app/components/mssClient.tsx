'use client'

import React, {JSX, useEffect, useState} from "react";
import {
    AssertArrayResult,
    DisplayInput, HandleJsonResponse, HandleTxtResponse, ImportDisplayInput,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString, ListPageItems, NewEntryInput,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    DisposedDisplay,
    ErrorDisplay,
    ParentDisplay,
    SpeciesArea,
    SubspeciesArea,
} from "@/app/components/formSubcomponents/commonClient";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {
    IsValidNote, NewEntryNotes,
    Note,
    NotesAreaInline
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
import {OnViewCreatorsQuadColArea} from "@/app/components/pcRunClient";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {WaterJarData} from "@/app/components/waterJarServer";
import {SporePrintRecentSelector} from "@/app/components/sporePrintClient";
import {WaterJarRecentSelector} from "@/app/components/waterJarClient";
import {LatestListDisplay} from "@/app/components/clientGeneric";
import {InlineEntry} from "@/app/components/agarRecipeClient";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import NotesArea from "@/app/components/formSubcomponents/notes";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";
import {LcSyringe} from "@/app/components/lcSyringeServer";

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
    //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    // const [perms, setPerms] = useState<EntryPerms | undefined>()
    const entriesCreatedDiv = ()=>{
        if(entriesCreated.length===0){
            return null
        }
        // TODO: STYLING SO NEW IS GREEN, newest is greener????
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
            // perms: perms, // TODO: validate on insert
        }
        subspecies && (body.subspecies = subspecies._id)
        notes.length>0 && (body.notes = notes)
        fetch(BaseExternalUrl+"/db/import/mss", {
            method: "POST",
            headers: {
                credentials: 'include',
                // 'Cookie': cookies, // TODO: may need
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleTxtResponse)
            .then((id) => {
                setEntriesCreated([...entriesCreated, id])
                // TODO: onCreate? redirect?
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
        <NotesArea readonly={false} updateParent={(ns) => { // TODO: notesFormArea?
            setNotes(ns.new.map((n) => {
                return n.data
            }))
        }}/>
        {/*<EntryPermsArea setEntryPerms={setPerms}/>*/}
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
                .then((entry) => { // TODO: MAKE DISPLAY UPDATES JUST RELOAD???
                    AssertMss(entry)
                    updateInitial(entry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: anything in here?
        ] // TODO: THIS!
        return <DisplayFormWrapper entryType={"mss"}>
            {/* TODO: TITLE? */}
            <ErrorDisplay err={err} headerLevel={headerLevel} />
            <ID id={data._id} txt={"Multispore Syringe"} entryType={"mss"} />
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel} canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
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
    // TODO: handle isTopLevel
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
        { sporePrintIn === undefined && <SporePrintRecentSelector  onSelect={setSporePrint}/>}{/* TODO: isTopLevel stuff. AllowCreate, etc*/}
        { waterJarIn === undefined && <WaterJarRecentSelector onSelect={setWaterJar} />} {/* TODO: isTopLevel stuff.  AllowCreate, etc*/}
        <NewEntryNotes setNotes={setNotes} />
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function MssInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<MssData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={data._id} txt={"Multispore Syringe"} entryType={"mss"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
            <SpeciesArea readonly={true} initial={data.species} />
            <SubspeciesArea readonly={true} currentSpecies={data.species} initialSub={data.subspecies}/>
            <SaleArea readonly={true} canCreateSale={false} sale={data.sale} />
            <DisposedDisplay readonly={true} disposed={data.disposed}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last updated: "} readonly={true} when={data.lastUpdated}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// export function MssListDisplay({data, onClick}: SingleListProps<MssData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <MssInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }

// TODO: MOVE! Also overhaul such that it can be closed
export function RecentSelectorV2<T>(
    {listUrlType, singleConstructor, assertion}:{
        listUrlType: string
        singleConstructor: (val: T, i: number)=>JSX.Element
        assertion: (a: any)=>void
}){
    const [data, setData] = useState<T[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(()=>{
        fetch(BaseExternalUrl + "/db/list/"+listUrlType, {
            method: "GET",
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        }).then((res) => {
            if(!res.ok){
                return res.text().then(txt=>{
                    throw new Error("response not ok: "+txt);
                })
            }
            res.json().then((resultData) => {
                AssertArrayResult<T>(resultData, assertion)
                setData(data)
            })
        }).catch(err1 => {
            console.log(JSON.stringify(err1))
            setErr(err1)
        })
    }, [])
    if (data.length === 0) {
        if (err !== undefined) {
            return <ErrorDisplay err={"failed to load mss selector: "+err}/>
        }
        return <div>{"Loading MSS selector"}</div>
    }
    // TODO: latestList should probably increment depth. It does at the time of writing this comment...
    return <LatestListDisplay data={data} constructor={singleConstructor}/> // TODO: LatestListDisplay does not currently close when selections occur.... Fix with new component for these selectors.
}

// TODO: HEAVILY TEST!!!! USE!!!!
export function MssRecentSelector({onSelect}:{onSelect:(selected?: MssData) => void}) {
    return <RecentSelectorV2<MssData> listUrlType={"mss"} assertion={AssertMss} singleConstructor={(val, i)=>{ // TODO: is this "msss"?
        return <MssInline data={val} expandByDefault={false} onClick={onSelect}/>
    }} />
}

export function MssListPageTable({data, onClick}: ListPageItems<MssData>) {
    const cols: ListTableColumn<MssData>[] = [
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
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}