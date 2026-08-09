'use client'

import React, {JSX, useContext, useState} from "react";
import {
    clientGetRequestHeaders,
    clientPostRequestHeaders, clientPostRequestHeadersMultipart,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoGetRequest,
    DoUpdateMultipartRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    importApiUrlFor,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
    RequiredKey, resolvePicsFormData, setFormFull, Subform,
    viewUrlFor,
} from "@/app/components/common";
import ReaderWriterSelector, {
    ReadRFIDButton,
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ErrorDisplay, MostRecentImageDisplay,
    ParentDisplay, PicsDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    IsValidNote, NewEntryNotes,
    Note, NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import DateArea from "@/app/components/formSubcomponents/date";
import {MssData} from "@/app/components/mssServer";
import ID from "@/app/components/formSubcomponents/id";
import {SpeciesData} from "@/app/components/speciesServer";
import {TransfersOutDisplay} from "@/app/components/transferClient";
import {SaleArea} from "@/app/components/saleClient";
import {AllEntries, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {SporePrintData, SporePrintSelectorCloseable} from "@/app/components/sporePrintServer";
import {WaterJarData, WaterJarSelectorCloseable} from "@/app/components/waterJarServer";
import {ExistingSpeciesSubspeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import TestAndValidate from "@/app/components/testing/untested";
import {AssertSporePrint} from "@/app/components/sporePrintClient";
import {AssertWaterJar} from "@/app/components/waterJarClient";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import {BaseExternalUrl} from "@/app/components/Constants";


export function AssertMss(input: any): asserts input is MssData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['subspecies', 'string'],
        ['parent', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('MSS assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['pics', IsValidPicWithNotesIncoming],
        ['transfersOut', IsString],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function MssImportDisplay({headerLevel}: ImportDisplayInput) { // TODO: USE ONLY FOR PURCHASED OR PRE-EXISTING MSS
    const [createdDate, setCreatedDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    // Non-required
    const [subspecies, setSubspecies] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])

    const [writeTagTo, setWriteTagTo] = useState<string | undefined>() // TODO: do we want this?
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
                return <EntryLinkForId key={i/* TODO: ensure ok*/} props={{displayId:b58id, linkId: b58id, entryType:"mss", openInNewTab: false}}/>// TODO: OPENINNEWTAB false ok?
            })}
        </div>
    }
    const tryImport = (e: React.MouseEvent) => {
        e.preventDefault()
        if(species===undefined){
            setErr("Species field cannot be undefined")
            return
        }
        const body: any = {
            creationDate: createdDate,
            species: species._id, // TODO: validate on insert
            // optional
            subspecies: subspecies,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        fetch(importApiUrlFor("mss"), { // TODO: use other func?
            method: "POST",
            headers: clientPostRequestHeaders,
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then(v => {
                AssertMss(v)
                window.location.assign(viewUrlFor("mss", v._id))
                // redirect(viewUrlFor("mss", newItem._id)) // TODO: del if working
            })
            .catch(ErrHandler(setErr));
    }
    return <ImportEntryFormWrapper entryType={"mss"}>
        <ErrorDisplay err={err}/>
        {entriesCreatedDiv()}
        <DateArea readonly={false} pre={"Created: "} when={Date.now()} updateParent={setCreatedDate}/>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {/*<SpeciesArea initial={species?._id} readonly={false} setSpecies={setSpecies}/>*/}
        {/*<SubspeciesArea initialSub={subspecies?._id} currentSpecies={species?._id} readonly={false} setSubspecies={setSubspecies}/>*/}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={tryImport}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export default function MssDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<MssData>) {
        const [initial, setInitial] = useState(data)

        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(data.pics))

        const [sale, setSale] = useState(data.sale)
        const [disposed, setDisposed] = useState(data.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL>(initial.acl)
        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const updateInitial= (updated: MssData)=>{
            setInitial(updated)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setTransfersOut(updated.transfersOut || [])
            setImages(InitialPicsEntries(updated.pics))
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const mssSubmit = () => {
            const formData = new FormData()
            const dataObj: any = {
                sale:sale,
                disposed:disposed,
                writeTagTo:writeTagTo,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            try {
                // Pics
                const picsInfo = resolvePicsFormData(images)
                const newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                setFormFull(formData, dataObj, newImages, undefined, undefined)
            } catch (caught: any) {
                console.log("error in submit")
                setErr(JSON.stringify(caught))
                return
            }
            DoUpdateMultipartRequest("mss",initial._id, formData, AssertMss, allCookies(cookies))
                .then(v=>{
                    updateInitial(new MssData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
            // DoUpdateRequest("mss",initial._id, dataObj, AssertMss, allCookies(cookies))
            //     .then(v=>{
            //         updateInitial(new MssData(v))
            //     })
            //     .catch(e=>{
            //         setErr("failed to update initial: "+JSON.stringify(e))
            //     })
        }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=> {
        const disp = initial.disposed !== undefined
        return !disp ? [
            WriteRfidOvcArea(initial._id),
        ]:[]
    }

        return <DisplayFormWrapper entryType={"mss"}>
            <ErrorDisplay err={err}/>
            <ID props={{id:data._id, txt:"Multispore Syringe", entryType:"mss"}}/>
            <MostRecentImageDisplay data={initial.mostRecentImage} showHeader={false}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} initialDisposed={initial.disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    <SaleArea sale={sale} setSale={setSale} readonly={readonly} canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <ParentDisplay parent={data.parent} parentType={"sporePrint"}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <TransfersOutDisplay thisId={data._id} thisEntryType={"mss"} transfersOut={data.transfersOut} allowNewTransferCreation={!readonly}  /*validTypesTo={["plate","slant","jar","bag"]} TODO: on go side*//>
            <PicsDisplay pix={initial.pics} readonly={readonly}
                         headerLevel={headerLevel} updateParent={setImages}/>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes} />
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                mssSubmit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
}

// this should only be used by the Spore Print Display Component
export function NewMssForm(
    {handlers,sporePrintIn,waterJarIn}: {handlers: NewEntryInput<MssData>, sporePrintIn?:SporePrintData, waterJarIn?:WaterJarData}){
    const [sporePrint, setSporePrint] = useState(sporePrintIn)
    const [waterJar, setWaterJar] = useState(waterJarIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
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
        const body: any = {
            sporePrintId: sporePrint._id,
            waterJar: waterJar._id, // TODO: DO THIS ON THE GO SIDE!
            notes: notes,
            writeTagTo: writeTagTo,
        }

        DoCreateRequest("mss", body, AssertMss, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(new MssData(v)) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }

    return <NewEntryFormWrapper entryType={"mss"}>
        <ErrorDisplay err={err}/>
        <TestAndValidate todos={["allow scans for Spore print or water jar selectors!"]}>
            { sporePrintIn === undefined && <div>
                <SporePrintSelectorCloseable  onSelect={setSporePrint} hideDisposed={true}/>
                <ReadRFIDButton handleTagRead={(tag)=>{
                    DoGetRequest("sporePrint", tag, AssertSporePrint, setErr).then(setSporePrint) // TODO: test and ensure ok!
                }} txt={"Or scan Spore Print RFID"}/>
            </div>}
        { waterJarIn === undefined && <div>
            <WaterJarSelectorCloseable doSelect={setWaterJar} creatorInPage={false} allowCreation={false} hideDisposed={true}/>
            <ReadRFIDButton handleTagRead={(tag)=>{
                DoGetRequest("waterJar", tag, AssertWaterJar, setErr).then(setWaterJar) // TODO: test and ensure ok!
            }} txt={"Or scan Water Jar RFID"}/>
        </div>}
        </TestAndValidate>
        <NewEntryNotes setNotes={setNotes} />
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

function getChildMssOf(parent?:string):Promise<MssData[]>{
    return new Promise((resolve,reject) => {
        reject("getChildMssOf not implemented yet!")
    })
    // TODO: FIX THIS WHOLE THING!
    // return fetch(BaseExternalUrl+'/db/listChildMss/'+parent, { // TODO: add endpoint!
    //     method: 'Get',
    //     credentials: 'include',
    //     headers: clientGetRequestHeaders,
    // }).then(res=> {
    //     return res.json()
    // }).then(json=>{
    //     try {
    //         return json as MssData[]
    //     } catch (e) {
    //         console.error(e)
    //         throw e
    //     }
    // }).catch(err=>{
    //     throw(err)
    // })
}

export function ChildMssArea({parent}:{parent?:string}){
    const [values, setValues] = useState<MssData[] | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const [collapsed, setCollapsed] = useState(false)
    const loadValues = (e: React.MouseEvent<HTMLButtonElement,MouseEvent>)=>{
        e.preventDefault()
        e.stopPropagation()
        getChildMssOf(parent).then((syringes:MssData[])=>{
            setValues(syringes)
            setErr(undefined)
        }).catch(e=>{
            setErr(JSON.stringify(e))
        })
    }
    const redirectToMss = (mss:MssData) => {
        window.location.assign(viewUrlFor("mss", mss._id)) // TODO: ENSURE OK!
    }
    const toggleCollapsed = (e: React.MouseEvent<HTMLButtonElement,MouseEvent>) => {
        e.preventDefault()
        e.stopPropagation()
        setCollapsed(!collapsed)
    }
    if(!parent){
        return null
    }
    if(values===undefined){
        return <Subform>
            <ErrorDisplay err={err}/>
            <button className={"basicButtonSmall"} onClick={loadValues}>{"Load child Spore Syringes"}</button>{/* TODO: ensure classes ok*/}
        </Subform>
    }
    if(values.length===0){
        // TODO: ensure if a mss is added and this is active it updates!
        return <Subform>{"No child syringes in database"}</Subform>
    }
    const toggleCollapsedButton = ()=>{
        return <button className={"basicButtonSmall"} onClick={toggleCollapsed}>{(collapsed?"Show":"Hide")+" child spore syringes"}</button>
    }
    const toggleButton = toggleCollapsedButton() // TODO: ensure ok and updates text when hiding/showing
    if(collapsed){
        return <Subform>
            {toggleButton}
        </Subform>
    }
    return <Subform>
        {toggleButton}
        <MssSelectorTable data={values} onClick={redirectToMss}/>
        {toggleButton}
    </Subform>
}

export function MssListPageTable({data, onClick, withLink}: ListPageItems<MssData>) {
    let cols: ListTableColumn<MssData>[] = [
        NewColumn("ID", (v)=>v._id, true),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Spec", (v)=>v.species||"", true),
        NewColumn("Subspec", v=>v.subspecies||"", true),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: MssData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new MssData(v)}}/>
}

export function MssSelectorTable({data, onClick}: ListPageItems<MssData>) {
    return <MssListPageTable data={data} onClick={onClick} withLink={true} />
}
export function MssSelector(
    {
        doSelect,
        allowCreate,
        hideDisposed = false
    }: {
        doSelect: (val: MssData | undefined) => void,
        allowCreate?: boolean,
        hideDisposed?:boolean
    }) {
    const table = (items: MssData[]):JSX.Element=>{
        return <MssSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"mss"} entryTypes={"mss"} doSelect={doSelect} asserter={AssertMss}
                                   table={table} hideDisposed={hideDisposed}>
        {allowCreate && <NewMssForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}