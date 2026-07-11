'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcData, LcSelectorCloseable} from "@/app/components/lcServer";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    clientPostRequestHeaders,
    ConfirmedCleanArea,
    ConfirmedCleanSelector,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoGetRequest, DoMultipartImportRequest, DoUpdateMultipartRequest,
    DoUpdateRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    importApiUrlFor,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
    RequiredKey, resolveContamsFormData, resolvePicsFormData, setFormFull,
    viewUrlFor,
} from "@/app/components/common";
import ReaderWriterSelector, {
    ReadRFIDButton,
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ErrorDisplay,
    GensFormDisplay, MostRecentImageDisplay,
    ParentDisplay, PicsDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {SpeciesData} from "@/app/components/speciesServer";
import ID from "@/app/components/formSubcomponents/id";
import {
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import {LcSyringeData} from "@/app/components/lcSyringeServer";
import {AllEntries, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {TransfersOutDisplay} from "@/app/components/transferClient";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import TestAndValidate from "@/app/components/testing/untested";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {AssertLc} from "@/app/components/lcClient";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import {AssertPlate} from "@/app/components/plateClient";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {PlateData} from "@/app/components/plateServer";

export function AssertLcSyringe(input: any): asserts input is LcSyringeData {
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
            throw new Error('Lc syringe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['sale', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['confirmedClean', 'boolean'],
        ['knownFruitable', 'boolean'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Lc syringe assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('LcSyringe assertion failure: required key ' + key + ' was not valid');
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
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Lc syringe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function LcSyringeImportDisplay() {
    // const cookies = useContext(CookiesContext)
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [confirmedClean, setConfirmedClean] = useState<boolean | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(1)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const ImportLcSyringe = () => {
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        const formData = new FormData()
        const dataObj: any = {
            creationDate: created,
            species: species._id,
            subspecies: subspecies,
            confirmedClean: confirmedClean,
            knownFruitable: knownFruitable,
            generation: generation,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        formData.set("data", JSON.stringify(dataObj))
        if (imageFile !== undefined) {
            formData.set("image", imageFile, "img")
        }
        DoMultipartImportRequest(formData, "lcSyringe", AssertLcSyringe, setErr, allCookies(cookies))

        // fetch(importApiUrlFor("lcSyringe"), {
        //     method: 'Post',
        //     body: JSON.stringify(dataObj),
        //     headers: clientPostRequestHeaders,
        // })
        //     .then(HandleJsonResponse)
        //     .then((newItem) => {
        //         AssertLcSyringe(newItem)
        //         window.location.assign(viewUrlFor("lcSyringe", newItem._id))
        //         // redirect(viewUrlFor("lcSyringe", newItem._id)) // TODO: del if working
        //     })
        //     .catch(ErrHandler(setErr));
    }
    return <ImportEntryFormWrapper entryType={"lcSyringe"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {/*<ExistingSpeciesSelector doSelect={setSpecies}/>*/}
        {/*<ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>*/}
        <ConfirmedCleanSelector updateParent={setConfirmedClean}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        <NewEntryNotes setNotes={setNotes}/>
        <div className={"centerH mt-2"}>
            <ImageSelector updateParent={setImageFile}/>
        </div>

        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>        <button className={"greenButton bottomButton"} onClick={ImportLcSyringe}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export default function LcSyringeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<LcSyringeData>) {
    const [initial, setInitial] = useState(data)
    const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(data.pics))

    const [transfersOut, setTransfersOut] = useState<string[]>(data.transfersOut || [])
    const [confirmedClean, setConfirmedClean] = useState<boolean | undefined>(data.confirmedClean)
    const [knownFruitable, setKnownFruitable] = useState(data.knownFruitable)
    const [disposed, setDisposed] = useState(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    const [acl, setAcl] = useState<ACL>(data.acl)
    const [err, setErr] = useState<string | undefined>()

    const updateInitial = (updated: LcSyringeData) => {
        setInitial(updated)
        setTransfersOut(updated.transfersOut || [])
        setConfirmedClean(updated.confirmedClean)
        setKnownFruitable(updated.knownFruitable)
        setDisposed(updated.disposed)
        setImages(InitialPicsEntries(updated.pics))
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
        setErr(undefined)
    }

    const cookies = useContext(CookiesContext)
    const lcSyringeSubmit = () => {
        const formData = new FormData()
        const dataObj: any = {
            confirmedClean: confirmedClean,
            knownFruitable: knownFruitable,
            disposed: disposed,
            sale: initial.sale, // TODO: this!?
            notes: notes,
            acl: MarshalAcl(acl),
        }
        try {
            // Pics
            const picsInfo = resolvePicsFormData(images)
            const newImages = picsInfo.images
            dataObj.images = picsInfo.obj
            // Set data on form
            setFormFull(formData, dataObj, newImages, undefined, undefined)
        } catch (caught: any) {
            console.log("error in submit")
            setErr(JSON.stringify(caught))
            return
        }
        console.log("submitting update request")
        DoUpdateMultipartRequest("lcSyringe",initial._id, formData, AssertLcSyringe, allCookies(cookies))
            .then(v=>{
                updateInitial(new LcSyringeData(v))
                console.log("updated initial state")
            })
            .catch(e=>{
                setErr("Error in parsing updated lcSyringe: "+JSON.stringify(e))
            })
        // DoUpdateRequest("lcSyringe",initial._id, dataObj, AssertLcSyringe, allCookies(cookies))
        //     .then(v=>{
        //         updateInitial(new LcSyringeData(v))
        //     })
        //     .catch(e=>{
        //         setErr(JSON.stringify(e))
        //     })
    }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=> {
        const disp = initial.disposed !== undefined
        return !disp ? [
            WriteRfidOvcArea(initial._id),
        ]:[]
    }
    return <DisplayFormWrapper entryType={"lcSyringe"}>
        <ErrorDisplay err={err}/>
        <ID props={{id:data._id, txt:"Liquid Culture Syringe", entryType:"lcSyringe"}}/>
        <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
        <MostRecentImageDisplay data={initial.mostRecentImage} showHeader={false}/>
        <FlexedArea>
            <FlexedSinglesGroup>
                <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                            initialDisposed={initial.disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                <ParentDisplay parent={initial.parent} parentType={"lc"}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <KnownFruitableArea initial={initial.knownFruitable} doSelect={setKnownFruitable} readonly={readonly}
                                    headerLevel={headerLevel}/>
                <ConfirmedCleanArea onSelect={setConfirmedClean} readonly={readonly} initial={initial.confirmedClean}
                                    headerLevel={headerLevel}/>
            </FlexedSinglesGroup>
        </FlexedArea>
        <TransfersOutDisplay thisId={initial._id} thisEntryType={"lcSyringe"} transfersOut={transfersOut}
                             allowNewTransferCreation={!readonly}/>
        <PicsDisplay pix={initial.pics} readonly={readonly}
                     headerLevel={headerLevel} updateParent={setImages}/>{/* Pics */}
        <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
        <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
            <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl} />
        </TogglableAreaWithDepth>
        {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
            e.stopPropagation();
            e.preventDefault()
            lcSyringeSubmit()
        }}>{"Update"}</button>}
    </DisplayFormWrapper>

}

export function NewLcSyringeForm({parentLc, onCreate, txt}: {
    parentLc?: LcData,
    onCreate?: (newItem: LcSyringeData) => void,
    txt: string
}) {
    const cookies = useContext(CookiesContext)
    const [itemsCreated, setItemsCreated] = useState<string[]>([])
    const [parent, setParent] = useState<LcData | undefined>(parentLc) // TODO: this ok to not call set??
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const createdItemsDiv = () => {
        if (itemsCreated.length === 0) {
            return null
        }
        return <div>
            <div>
                <div>{"Lc syringes Created:"}</div>
            </div>
            {itemsCreated.map((createdLc) => {
                const b58id = createdLc
                return <EntryLinkForId key={createdLc/* TODO: ensure ok*/} props={{displayId: b58id, linkId: b58id, entryType: "lcSyringe", openInNewTab: false /* TODO: ok?*/}}/>
            })}
        </div>
    }

    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!parent) {
            setErr("A parent must be selected")
            return
        }
        const body: any = {
            writeTagTo: writeTagTo,
            parent: parent._id,
            notes: notes,
        }
        DoCreateRequest("lcSyringe", body, AssertLcSyringe, allCookies(cookies))
            .then(v=>{
                onCreate ? onCreate(v) : console.log("no onCreate function provided")
                setItemsCreated([...itemsCreated, v._id])
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }

    return <NewEntryFormWrapper entryType={"lcSyringe"}>
            <div>{txt}</div>
            {createdItemsDiv()}
            <ErrorDisplay err={err}/>
        {!parent && <div>
            <LcSelectorCloseable doSelect={setParent} hideDisposed={true}/>
            <ReadRFIDButton handleTagRead={(val)=>{
                DoGetRequest("lc", val, AssertLc, setErr).then(setParent)
            }} txt={"Or read parent LC's RFID tag"}/>
        </div>}
            <NewEntryNotes setNotes={setNotes}/>
            <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
            <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function LcSyringeListPageTable({data, onClick, withLink}: ListPageItems<LcSyringeData>) {
    let cols: ListTableColumn<LcSyringeData>[] = [
        NewColumn("ID", (v)=>v._id, true),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Spec", (v)=>v.species||"", true),
        NewColumn("Subspec", v=>v.subspecies||"", true),
        NewColumn("Clean",v=>v.confirmedClean?(v.confirmedClean?"clean":"dirty"):"?", true),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: LcSyringeData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new LcSyringeData(v)}}/>
}
export function LcSyringeSelectorTable({data, onClick}: ListPageItems<LcSyringeData>) {
    return <LcSyringeListPageTable data={data} onClick={onClick} withLink={true} />
}

export function LcSyringeSelector(
    {
        doSelect,
        hideDisposed = false
    }: {
        doSelect: (val?: LcSyringeData) => void,
        hideDisposed?: boolean
    }) {
    const table = (items: LcSyringeData[]):JSX.Element=>{
        return <LcSyringeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"lcSyringe"} entryTypes={"lcSyringes"} doSelect={doSelect} asserter={AssertLcSyringe}
                                   table={table} hideDisposed={hideDisposed}/>
}
