'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    SlantData
} from "@/app/components/slantServer";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateMultipartRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup, ImportDisplayInput, ImportEntryFormWrapper, ListPageItems,
    ListPageTable,
    ListTableColumn,
    MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey, RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData,
    setFormData,
    setFormImages,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ErrorDisplay,
    GensFormDisplay, MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {
    AgarBatchArea,
} from "@/app/components/agarBatchClient";
import {
    ContaminationForm, ContamsDisplay, InitialContamState, IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {AgarBatchData, AgarBatchSelectorCloseable} from "@/app/components/agarBatchServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import { SaleArea} from "@/app/components/saleClient";
import {
    ExistingSpeciesSelector,
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {
    AclDisplay,
    IsValidAcl,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import TestAndValidate from "@/app/components/testing/untested";

export function AssertSlant(input: any): asserts input is SlantData {
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
            throw new Error('Slant assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['agarBatch', 'string'],
        ['stickType','string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Slant assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Slant assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Slant assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Slant assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function SlantImportDisplay({headerLevel}:ImportDisplayInput) {
    const [created, setCreated] = useState<number>(Date.now())
    const [stickType, setStickType] = useState<string | undefined>(undefined)
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<string | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const ImportSlant = () => {
        const formData = new FormData()
        formData.set('data', JSON.stringify({
            creationDate:created, // TODO: validate not in future or too far in the past (do on all imports)
            stickType: stickType,
            // Optional
            species: species?._id,
            subspecies: subspecies,
            knownFruitable: knownFruitable,
            generation: generation,
            writeTagTo: writeTagTo,
        }))
        if(imageFile!==undefined){
            formData.set("image", imageFile, "imgFile")
        }

        MultipartImportRequest(formData, "slant", AssertSlant, setErr, allCookies(cookies))
    }
    return <ImportEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {/*<ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>*/}
        {/*{species?<ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel}/>:null}*/}
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <SlantStickSelector setStickType={setStickType}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={ImportSlant}>{"Import Slant"}</button>
    </ImportEntryFormWrapper>
}

export default function SlantDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<SlantData>) {
        const [initial, setInitial] = useState(data)
        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const updateInitial = (updated: SlantData)=>{
            setInitial(updated)
            setImages(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const slantSubmit = ()=>{
            const formData = new FormData()
            const dataObj:any={
                knownFruitable:knownFruitable,
                sale: sale,
                disposed:disposed,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            try {
                // Pics
                const picsInfo = resolvePicsFormData(images)
                const newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                const contamsInfo = resolveContamsFormData(contams)
                const newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(formData, dataObj)
                setFormImages(formData, "newPic", newImages)
                setFormImages(formData, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }
            DoUpdateMultipartRequest("slant",initial._id, formData, AssertSlant, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SlantData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        // TODO: DIFFERENTIATE BETWEEN UNINNOCULATED AND INNOCULATED DISPLAY
        const ovcs: OnViewCreatorQuadCol[] = [
            WriteRfidOvcArea(initial._id),
            // TODO: CREATE FRUIT
            // TODO: spore print?
            // TODO: spore swab?
        ]
        const innoculated = initial.species // TODO: USE THIS!
        return (
            <DisplayFormWrapper entryType={"slant"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID props={{id:data._id, txt:"Slant", entryType:"slant"}}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} initialDisposed={initial.disposed}
                                                    readonly={readonly} setDisposedOnParent={setDisposed}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>
                        <AgarBatchArea agarBatchId={initial.agarBatch} headerLevel={headerLevel}/>
                        <div>
                            {"Stick type:"+(initial.stickType || "none")}
                        </div>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <InnocDisplay innoc={initial.innoc}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                                  canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedArea>

                <TransfersOutDisplay thisId={initial._id} thisEntryType={"slant"} transfersOut={transfersOut} allowNewTransferCreation={!readonly}/>
                <PicsDisplay pix={initial.pics || []} updateParent={setImages} readonly={readonly} headerLevel={headerLevel}/>{/* Pics */}
                <ContamsDisplay initial={initial.contamination || []} updateParent={setContams} readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    slantSubmit()
                }}>{"Update"}</button>}

            </DisplayFormWrapper>
        )
}

export function NewSlantForm({handlers,agarBatchIn}: {handlers: NewEntryInput<SlantData>, agarBatchIn?:AgarBatchData}){
    const [agarBatch, setAgarBatch] = useState(agarBatchIn)
    const [stickType, setStickType] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
    const createSlant = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(agarBatch===undefined){
            setErr("An agar batch must be selected")
            return
        }
        // agarBatches always are created with pcRuns, so there is no need to ensure the batch has a run beforehand...
        const body: any = {
            agarBatch: agarBatch._id,
            stickType: stickType,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("slant", body, AssertSlant, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err}/>
        <SlantStickSelector setStickType={setStickType}/>
        <TestAndValidate todos={["ensure providing agarBatch makes selector disappear"]}>
        {agarBatchIn===undefined && <AgarBatchSelectorCloseable doSelect={setAgarBatch} allowCreation={handlers.isTopLevel} creatorInPage={handlers.isTopLevel/* TODO: ok?*/}/>}
        </TestAndValidate>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createSlant} onSubmit={(e)=>{e.preventDefault();}}>{"Create"}</button>
    </NewEntryFormWrapper>
}

// TODO: validate works properly
export function SlantStickSelector(x: { setStickType: (s?: string) => void }){
    return <div>
        {"Slant stick: "}<select className={"tailwindSelector"} defaultValue={"none"} disabled={false} onChange={(e) => {x.setStickType((e.currentTarget.value==="none")?undefined:e.currentTarget.value)}}>
        {["none","tongueDepressor"].map((s, i: number) => {
            return <option value={s} key={i}>{s}</option>
        })}
    </select>
    </div>
}

export function SlantListPageTable({data, onClick, withLink}: ListPageItems<SlantData>) {
    let cols: ListTableColumn<SlantData>[] = [
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
        cols = [...cols, NewColumn("Link", (v: SlantData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SlantData(v)}}/>
}
export function SlantSelectorTable({data, onClick}: ListPageItems<SlantData>) {
    return <SlantListPageTable data={data} onClick={onClick} withLink={true} />
}
export function SlantSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: SlantData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: SlantData[]):JSX.Element=>{
        return <SlantSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"slant"} entryTypes={"slants"} doSelect={doSelect} asserter={AssertSlant}
                                   table={table}>
        {allowCreate && <NewSlantForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}