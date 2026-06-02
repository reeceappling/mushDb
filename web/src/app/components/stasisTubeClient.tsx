'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    Data,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    ImportDisplayInput, ImportEntryFormWrapper,
    ListPageItems, ListPageTable, ListTableColumn, MultipartImportRequest, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
    resolveContamsFormData,
    resolvePicsFormData, setFormData,
    setFormImages,
} from "@/app/components/common";
import ReaderWriterSelector, {WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    ErrorDisplay,
    GensFormDisplay, MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {
    ContaminationForm, ContamsDisplay, InitialContamState, IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertStasisTube(input: any): asserts input is StasisTubeData {
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
            throw new Error('StasisTube assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['waterSource', 'string'],
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
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('StasisTube assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('StasisTube assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('StasisTube assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function StasisTubeImportDisplay() {
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const importEntry = () => {
        let formData = new FormData()
        let dataObj: any = {
            created:created,
        }
        if(species===undefined){
            setErr("Species must be set!")
            return
        }
        dataObj.species = species._id
        subspecies && (dataObj.subspecies = subspecies._id)
        knownFruitable && (dataObj.knownFruitable = knownFruitable)
        generation && (dataObj.generation = generation)
        if(imageFile!==undefined){
            formData.set("image", imageFile, "imgFile")
        }
        writeTagTo && (dataObj.writeTagTo=writeTagTo)

        MultipartImportRequest(formData, "stasisTube", AssertStasisTube, setErr, allCookies(cookies))
        // TODO: reenable if does not work without cookies... SendMultipartRequest(importApiUrlFor("stasisTube"), cookies, formData)
        // SendMultipartRequest2(importApiUrlFor("stasisTube"), formData)
        //     .then(HandleJsonResponse)
        //     .then(newItem => {
        //         AssertStasisTube(newItem)
        //         redirect(viewUrlFor("stasisTube", newItem._id))
        //     })
        //     .catch(ErrHandler(setErr));
    }
    return <ImportEntryFormWrapper entryType={"stasisTube"}>
        {err!=undefined && <div>{"Error: "+err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSelector doSelect={setSpecies}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} />
        <button className={"greenButton"} onClick={importEntry}>{"Import Stasis Tube"}</button>
    </ImportEntryFormWrapper>
}

export default function StasisTubeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertStasisTube(data)
        const [initial, setInitial] = useState(data)
        const existingNotes: Note[] = initial.notes || []
        const initNotes: Data<Note>[] = existingNotes.map((n) => {
            return {data: n, disabled: false}
        })

        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [sale, setSale] = useState(initial.sale)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing:initNotes,new:[]})
        // State helpers
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: StasisTubeData)=>{
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
        }
        const cookies = useContext(CookiesContext)
        const stasisTubeSubmit = () => {
            let formData = new FormData()
            let dataObj:any={
                knownFruitable: knownFruitable,
                sale: sale,
                disposed: disposed,
                notes: notes,
                acl: acl,
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(images)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(formData, dataObj)
                setFormImages(formData, "newPic", newImages)
                setFormImages(formData, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateRequest("stasisTube",data._id, formData, AssertStasisTube, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))

            // SendMultipartRequest(updateApiUrlFor("stasisTube", initial._id), cookies, formData)
            //     .then(HandleJsonResponse)
            //     .then((entry) => {
            //         AssertStasisTube(entry)
            //         updateInitial(entry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            WriteRfidOvcArea(initial._id),
        ]
        return (
            <DisplayFormWrapper entryType={"stasisTube"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Stasis Tube"} entryType={"stasisTube"}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel} />
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <PcRunArea binaryId={initial.pcRun} headerLevel={headerLevel}/>
                        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly} headerLevel={headerLevel}/>
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel} canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore} headerLevel={headerLevel} />

                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        <InnocDisplay innoc={initial.innoc} openInNewTab={false}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <TransfersOutDisplay thisId={initial._id} thisEntryType={"stasisTube"} allowNewTransferCreation={!readonly} transfersOut={transfersOut} validTypesTo={["plate","stasisTube","jar"/* TODO: ANYMORE????*/]} headerTxt={"Transfers"}/>
                <PicsDisplay pix={initial.pics || []} updateParent={setImages} readonly={readonly} headerLevel={headerLevel} />{/* Pics */}
                <ContamsDisplay initial={initial.contamination || []} updateParent={setContams} readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    stasisTubeSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: StasisTube data format incorrect: " + err}</div>
    }
}

export function NewStasisTubeForm({handlers, pcRunIn}: {handlers: NewEntryInput<StasisTubeData>, pcRunIn?: PcRunData}){
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const createStasisTube = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(pcRun===undefined){
            setErr("pc run must be defined")
            return
        }
        let body:any={
            pcRun: pcRun._id,
            notes: notes,
            writeTagTo:writeTagTo,
        }
        DoCreateRequest("stasisTube", body, AssertStasisTube, allCookies(cookies))
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("stasisTube"), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body),
        // })
        //     .then(HandleJsonResponse)
        //     .then((entry) => {
        //         AssertStasisTube(entry)
        //         handlers.onCreate && handlers.onCreate(entry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    return <NewEntryFormWrapper entryType={"stasisTube"}>
        <ErrorDisplay err={err} />
        {pcRunIn !== undefined && <PcRunSelectorCloseable doSelect={setPcRun} allowCreation={handlers.isTopLevel} creatorInPage={true}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} />
        <button className={"greenButton"} onClick={createStasisTube}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function StasisTubeListPageTable({data, onClick, withLink}: ListPageItems<StasisTubeData>) {
    let cols: ListTableColumn<StasisTubeData>[] = [
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
        cols = [...cols, NewColumn("Link", (v: StasisTubeData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function StasisTubeSelectorTable({data, onClick}: ListPageItems<StasisTubeData>) {
    return <StasisTubeListPageTable data={data} onClick={onClick} withLink={true} />
}
export function StasisTubeSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: StasisTubeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: StasisTubeData[]):JSX.Element=>{
        return <StasisTubeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"stasisTube"} entryTypes={"stasisTubes"} doSelect={doSelect} asserter={AssertStasisTube}
                                   table={table}>
        {allowCreate && <NewStasisTubeForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}